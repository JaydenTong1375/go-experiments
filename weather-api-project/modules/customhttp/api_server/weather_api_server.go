package apiserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	apiClient "weather-api-project/modules/customhttp/api_client"
	redis "weather-api-project/modules/redis"

	"github.com/google/uuid"
)

const dbTableWeatherDays = "weather_days_record"
const dbTableWeatherHours = "weather_hours_record"

func apiGetWeatherData(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	// Decode JSON body into a dynamic struct (get data from body)
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	country, bCountry := body["country"].(string)
	searchDate, bSearchDate := body["date"].(string)

	if !bCountry {
		writeResultJSON(w, "Unable to find country in body", http.StatusBadRequest)
		return
	}

	if !bSearchDate {
		writeResultJSON(w, "Unable to find startDate in body", http.StatusBadRequest)
		return
	}

	if db == nil {
		writeResultJSON(w, "db is invalid", http.StatusInternalServerError)
		return
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//If the relevant data can't be found in the database, then fetch it either from the Redis cache or from the official website
	redisKey := fmt.Sprintf("%s:%s", country, searchDate)

	redisClient, redisErr := redis.GetRedisClient()

	if redisErr != nil {
		writeResultJSON(w, fmt.Sprintf("Unable to get redis client %v", redisErr), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	redisResult, getErr := redisClient.Get(ctx, redisKey).Result()

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//If the relevant data cannot be found in Redis, fetch the latest weather data from the official website or local DB
	if getErr != nil {

		///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		//Check the database first; if the relevant data doesn't exist, fetch the latest data from the official website.
		checkQuery := fmt.Sprintf(`
		SELECT wdr.day_id, wdr.country, wdr.date , wdr.timezone, wdr.conditions, wdr.description,
		CONCAT("[",
			GROUP_CONCAT(
			CONCAT("{", '"time": "', whr.time, '", "conditions": "' ,whr.conditions,'", "stations": ', whr.stations, "}")
			ORDER BY whr.time ASC
			SEPARATOR ' , '
			)
		,"]") AS hours
		FROM %s wdr
		INNER JOIN (
			SELECT *
			FROM %s AS t1
			WHERE NOT EXISTS (
				SELECT 1 FROM %s AS t2
				WHERE t2.day_id = t1.day_id
				  AND t2.time = t1.time
				  AND t2.created_at > t1.created_at
			)
		) AS whr ON wdr.day_id = whr.day_id
		WHERE wdr.country = ? AND wdr.date = ?
		GROUP BY wdr.day_id`, dbTableWeatherDays, dbTableWeatherHours, dbTableWeatherHours)

		var queryDayID, queryCountry, queryDate, queryTimezone, queryConditions, queryDes, queryHours string

		scanErr := db.QueryRow(checkQuery, country, searchDate).Scan(&queryDayID, &queryCountry, &queryDate, &queryTimezone, &queryConditions, &queryDes, &queryHours)

		if scanErr != nil {
			log.Printf("failed to scan %v \n", scanErr)
		}

		if queryDayID != "" {
			var dayRes weatherRecordDay
			dayRes.DayID = queryDayID
			dayRes.Country = queryCountry
			dayRes.Date = queryDate
			dayRes.Timezone = queryTimezone
			dayRes.Conditions = queryConditions
			dayRes.Description = queryDes

			var hrsRes []weatherRecordHours
			unmarshalHrsErr := json.Unmarshal([]byte(queryHours), &hrsRes)

			if unmarshalHrsErr != nil {
				writeResultJSON(w, fmt.Sprintf("failed to unmarshal hours data %v", unmarshalHrsErr), http.StatusInternalServerError)
				return
			}

			dayRes.Hours = hrsRes

			var res response
			res.Message = "fetch from local db"
			res.Result = dayRes

			jsonRes, marshalJsonResErr := json.Marshal(res)

			if marshalJsonResErr != nil {
				writeResultJSON(w, fmt.Sprintf("failed to marshal jsonRes %v", marshalJsonResErr), http.StatusInternalServerError)
				return
			}

			writeResultJSON(w, string(jsonRes), http.StatusOK)
			return
		}

		///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		//If the relevant data cannot be found in local DB, fetch the latest weather data from the official website,
		//extract the necessary information, and save it to local DB.
		res, resErr := apiClient.GetWeatherDataFromOfficialWeb(country, searchDate)

		if resErr != nil {
			writeResultJSON(w, fmt.Sprintf("Unable to get weather data from official web %v", resErr), http.StatusInternalServerError)
			return
		}

		//Extract the relevant data
		var unmarshalRes map[string]interface{}
		unmarshalResErr := json.Unmarshal([]byte(res), &unmarshalRes)

		if unmarshalResErr != nil {
			writeResultJSON(w, fmt.Sprintf("Unable to unmarshal weather data %v\n %s", unmarshalResErr, res), http.StatusInternalServerError)
			return
		}

		var weatherData []weatherRecordDay

		timezone := unmarshalRes["timezone"].(string)

		for _, daysRawData := range unmarshalRes["days"].([]interface{}) {

			//EXtract days data
			daysData := daysRawData.(map[string]interface{})

			date := daysData["datetime"].(string)
			conditions := daysData["conditions"].(string)
			description := daysData["description"].(string)

			dayID := uuid.New().String()

			var dayRecord weatherRecordDay
			dayRecord.DayID = dayID
			dayRecord.Country = country
			dayRecord.Timezone = timezone
			dayRecord.Date = date
			dayRecord.Conditions = conditions
			dayRecord.Description = description

			//Extract hours data
			var weatherHoursData []weatherRecordHours

			for _, hoursRawData := range daysData["hours"].([]interface{}) {

				hoursData := hoursRawData.(map[string]interface{})

				time := hoursData["datetime"].(string)
				conditions := hoursData["conditions"].(string)
				stations := hoursData["stations"].([]interface{})

				var hoursRecord weatherRecordHours
				hoursRecord.DayID = dayID
				hoursRecord.Time = time
				hoursRecord.Conditions = conditions

				for _, s := range stations {
					hoursRecord.Stations = append(hoursRecord.Stations, s.(string))
				}

				weatherHoursData = append(weatherHoursData, hoursRecord)
			}

			dayRecord.Hours = weatherHoursData

			weatherData = append(weatherData, dayRecord)

		}

		jsonDayRecord, dayMarshalErr := json.MarshalIndent(weatherData[0], "", "  ")

		if dayMarshalErr != nil {
			writeResultJSON(w, fmt.Sprintf("Unable to marshal response %v", dayMarshalErr), http.StatusInternalServerError)
			return
		}

		//Save the latest weather data fetched from the official website to the redis 🔔
		saveErr := redisClient.Set(ctx, redisKey, string(jsonDayRecord), 0).Err()

		if saveErr != nil {
			writeResultJSON(w, fmt.Sprintf("Unable to save weather data to redis %v", saveErr), http.StatusInternalServerError)
			return
		}

		//Save the latest weather data fetched from the official website to the database 🔔
		tx, beginErr := db.Begin()
		if beginErr != nil {
			writeResultJSON(w, fmt.Sprintf("Failed to begin transaction: %v", beginErr), http.StatusInternalServerError)
			return
		}

		for _, wd := range weatherData {
			saveNewDaysQuery := fmt.Sprintf(`
			INSERT INTO %s (day_id, country, timezone, date, conditions, description) VALUES (?, ?, ?, ?, ?, ?)
			`, dbTableWeatherDays)

			_, daysExecErr := tx.Exec(saveNewDaysQuery, wd.DayID, wd.Country, wd.Timezone, wd.Date, wd.Conditions, wd.Description)

			if daysExecErr != nil {
				log.Printf("failed to save days into db %v \n", daysExecErr)
				continue
			}

			for _, hr := range wd.Hours {

				saveNewHoursQuery := fmt.Sprintf(`
				INSERT INTO %s (day_id, time, conditions, stations) VALUES (?, ?, ?, ?)
				`, dbTableWeatherHours)

				jsonStation, marshalStationErr := json.Marshal(hr.Stations)

				if marshalStationErr != nil {
					log.Printf("failed to marshal Stations %v \n", marshalStationErr)
					continue
				}

				_, hoursExecErr := tx.Exec(saveNewHoursQuery, hr.DayID, hr.Time, hr.Conditions, string(jsonStation))

				if hoursExecErr != nil {
					log.Printf("failed to save hours into db %v \n", hoursExecErr)
					continue
				}
			}

		}

		// Commit transaction only if all queries succeed
		if commitErr := tx.Commit(); commitErr != nil {
			writeResultJSON(w, fmt.Sprintf("Failed to commit transaction: %v", commitErr), http.StatusInternalServerError)
			return
		}

		var res2 response
		res2.Message = "fetch from offical web"
		res2.Result = weatherData[0]

		jsonRes2, marshalJsonRes2Err := json.Marshal(res2)

		if marshalJsonRes2Err != nil {
			writeResultJSON(w, fmt.Sprintf("failed to marshal jsonRes %v", marshalJsonRes2Err), http.StatusInternalServerError)
			return
		}

		//get the latest weather data from the official web
		writeResultJSON(w, string(jsonRes2), http.StatusOK)
		return
	}

	//get the weather data from the redis caches
	log.Printf("\n\n\n Redis result -> %s \n\n\n", redisResult)
	var redisUnmarshalData weatherRecordDay
	unmarshalErr := json.Unmarshal([]byte(redisResult), &redisUnmarshalData)

	if unmarshalErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to unmarshal redisResult %v", unmarshalErr), http.StatusInternalServerError)
		return
	}

	var res3 response
	res3.Message = "fetch from redis cache"
	res3.Result = redisUnmarshalData

	jsonRes3, marshalJsonRes3Err := json.Marshal(res3)

	if marshalJsonRes3Err != nil {
		writeResultJSON(w, fmt.Sprintf("failed to marshal jsonRes %v", marshalJsonRes3Err), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(jsonRes3), http.StatusOK)
}

func apiUpdateWeatherData(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	if db == nil {
		writeResultJSON(w, "db is invalid", http.StatusInternalServerError)
		return
	}

	// Decode JSON body into a dynamic struct (get data from body)
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResultJSON(w, fmt.Sprintf("Invalid JSON %v", err), http.StatusBadRequest)
		return
	}

	dayID := body["day_id"].(string)
	country := body["country"].(string)
	timezone := body["timezone"].(string)
	date := body["date"].(string)
	conditions := body["conditions"].(string)
	description := body["description"].(string)
	hours := body["hours"].([]interface{})

	tx, beginErr := db.Begin()

	if beginErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to begin transaction: %v", beginErr), http.StatusInternalServerError)
		return
	}

	//Update day in DB
	updateDayQuery := fmt.Sprintf(`
	UPDATE %s SET country = ?, timezone = ?, date = ?, conditions = ?, description = ?
	WHERE day_id = ?`, dbTableWeatherDays)

	updateDaysRes, updateDayErr := tx.Exec(updateDayQuery, country, timezone, date, conditions, description, dayID)

	if updateDayErr != nil {
		tx.Rollback()
		writeResultJSON(w, fmt.Sprintf("Failed to insert data: %v", updateDayErr), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := updateDaysRes.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		writeResultJSON(w, "No matching record found for the provided day_id.", http.StatusBadRequest)
		return
	}

	//Update hours in DB
	for _, h := range hours {
		hd := h.(map[string]interface{})
		time := hd["time"].(string)
		conditions := hd["conditions"].(string)
		stations := hd["stations"].([]interface{})

		log.Printf("time: %s | conditions: %s", time, conditions)

		jsonStations, stationMarshalErr := json.Marshal(stations)

		if stationMarshalErr != nil {
			log.Printf("Failed to marshal stations: %v\n", stationMarshalErr)
			continue
		}

		insertHoursQuery := fmt.Sprintf(`
		INSERT INTO %s (day_id, time, conditions, stations) VALUES (?, ?, ?, ?)`, dbTableWeatherHours)

		_, insertHoursErr := tx.Exec(insertHoursQuery, dayID, time, conditions, string(jsonStations))

		if insertHoursErr != nil {
			log.Printf("Failed to insert data: %v\n", insertHoursErr)
			continue
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		writeResultJSON(w, fmt.Sprintf("Failed to commit transaction: %v", commitErr), http.StatusInternalServerError)
		return
	}

	//Get the latest data from the local DB and store it in the Redis cache
	fetchQuery := fmt.Sprintf(`
		SELECT wdr.day_id, wdr.country, wdr.date , wdr.timezone, wdr.conditions, wdr.description,
		CONCAT("[",
			GROUP_CONCAT(
			CONCAT("{", '"time": "', whr.time, '", "conditions": "' ,whr.conditions,'", "stations": ', whr.stations, "}")
			ORDER BY whr.time ASC
			SEPARATOR ' , '
			)
		,"]") AS hours
		FROM %s wdr
		INNER JOIN (
			SELECT *
			FROM %s AS t1
			WHERE NOT EXISTS (
				SELECT 1 FROM %s AS t2
				WHERE t2.day_id = t1.day_id
				  AND t2.time = t1.time
				  AND t2.created_at > t1.created_at
			)
		) AS whr ON wdr.day_id = whr.day_id
		WHERE wdr.country = ? AND wdr.date = ?
		GROUP BY wdr.day_id`, dbTableWeatherDays, dbTableWeatherHours, dbTableWeatherHours)

	var queryDayID, queryCountry, queryDate, queryTimezone, queryConditions, queryDes, queryHours string

	scanErr := db.QueryRow(fetchQuery, country, date).Scan(&queryDayID, &queryCountry, &queryDate, &queryTimezone, &queryConditions, &queryDes, &queryHours)

	if scanErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to scan %v", scanErr), http.StatusInternalServerError)
		return
	}

	if queryDayID == "" {
		writeResultJSON(w, "failed to get the latest data from local db.", http.StatusInternalServerError)
		return
	}

	var dayRes weatherRecordDay
	dayRes.DayID = queryDayID
	dayRes.Country = queryCountry
	dayRes.Date = queryDate
	dayRes.Timezone = queryTimezone
	dayRes.Conditions = queryConditions
	dayRes.Description = queryDes

	var hrsRes []weatherRecordHours
	unmarshalHrsErr := json.Unmarshal([]byte(queryHours), &hrsRes)

	if unmarshalHrsErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to unmarshal hours data %v", unmarshalHrsErr), http.StatusInternalServerError)
		return
	}

	dayRes.Hours = hrsRes

	redisKey := fmt.Sprintf("%s:%s", country, date)

	redisClient, redisErr := redis.GetRedisClient()

	if redisErr != nil {
		writeResultJSON(w, fmt.Sprintf("Unable to get redis client %v", redisErr), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	marshalDayRes, marshalDayErr := json.Marshal(dayRes)

	if marshalDayErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to marshal dayRes. %v", marshalDayErr), http.StatusOK)
		return
	}

	saveReditErr := redisClient.Set(ctx, redisKey, string(marshalDayRes), 0).Err()

	if saveReditErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to save data to redis. %v", saveReditErr), http.StatusOK)
		return
	}

	var res response
	res.Message = "Weather data updated successfully. Fetched from local database."
	res.Result = dayRes

	jsonRes, marshalJsonResErr := json.Marshal(res)

	if marshalJsonResErr != nil {
		writeResultJSON(w, fmt.Sprintf("failed to marshal jsonRes %v", marshalJsonResErr), http.StatusInternalServerError)
		return
	}

	writeResultJSON(w, string(jsonRes), http.StatusOK)
}
