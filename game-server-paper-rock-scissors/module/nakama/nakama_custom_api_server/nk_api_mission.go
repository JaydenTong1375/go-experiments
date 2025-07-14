package nakamacustomapiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

const dbTableUserMission = "users_mission"
const dbTableMissionTracker = "users_mission_tracker"
const dbTableObjectiveTracker = "mission_objectives_tracker"

func apiSaveUserQuest(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)
	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
		return
	}

	//Get user id
	userID, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	tx, txErr := sqlDB.Begin()

	if txErr != nil {
		writeResultJSON("❌ Failed to start a transaction.", http.StatusInternalServerError, w)
		return
	}

	defer tx.Rollback()

	//Get payload from body
	var payload map[string]interface{}

	DecodeErr := json.NewDecoder(r.Body).Decode(&payload)
	if DecodeErr != nil {
		writeResultJSON("Invalid JSON body", http.StatusBadRequest, w)
		return
	}

	payloadQuests, bIsQuestsValid := payload["quests"].(string)

	if bIsQuestsValid == false {
		writeResultJSON("quests not found in payload", http.StatusBadRequest, w)
		return
	}

	type Quest struct {
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Type        string                   `json:"type"`
		Objective   []map[string]interface{} `json:"objective"`
	}

	var quest []Quest
	unmarshalQuestErr := json.Unmarshal([]byte(payloadQuests), &quest)

	if unmarshalQuestErr != nil {
		writeResultJSON(fmt.Sprintf("failed to unmarshal quest data. %v", unmarshalQuestErr), http.StatusInternalServerError, w)
		return
	}

	numberOfQuest := len(quest)
	AddedQuestCount := 0

	for _, q := range quest {

		missionID := uuid.New().String()

		// Add a new mission to the user's mission tracker table
		insertMissionTrackerQuery := fmt.Sprintf(`INSERT INTO %s (user_id, mission_id, mission_name, mission_type, mission_description) VALUES (?, ?, ?, ?, ?)`, dbTableMissionTracker)

		_, insertErr := tx.Exec(insertMissionTrackerQuery, userID, missionID, q.Name, q.Type, q.Description)

		if insertErr != nil {
			log.Printf("Failed to add a new quest to the user's mission tracker. %v \n", insertErr)
			continue
		}

		nmberOfObjectives := len(q.Objective)
		AddedObjectiveCount := 0

		// Add objectives record to the user's objectives tracker
		for _, o := range q.Objective {

			objectiveID := uuid.New().String()

			objectiveName, bIsOrderValid := o["name"]

			objectiveOrder, bIsNameValid := o["order"]

			objectiveDescription, bIsDescriptionValid := o["description"]

			if bIsNameValid == false {
				log.Printf("Objective: name not found\n")
				continue
			}

			if bIsOrderValid == false {
				log.Printf("Objective: order not found\n")
				continue
			}

			if bIsDescriptionValid == false {
				log.Printf("Objective: description not found\n")
				continue
			}

			insertObjectiveQuery := fmt.Sprintf(`INSERT INTO %s (mission_id, objective_id, objective_order, objective_name, objective_description) VALUES (?, ?, ?, ?, ?)`, dbTableObjectiveTracker)

			_, insertObjectErr := tx.Exec(insertObjectiveQuery, missionID, objectiveID, objectiveOrder, objectiveName, objectiveDescription)

			if insertObjectErr != nil {
				log.Printf("Failed to add objectives to objectives tracker. %v \n", insertObjectErr)
				continue
			}

			AddedObjectiveCount += 1
		}

		if AddedObjectiveCount < nmberOfObjectives {
			continue
		}

		AddedQuestCount += 1
	}

	if AddedQuestCount < numberOfQuest {
		writeResultJSON(fmt.Sprintf("❌ Failed to add some quests. Total quests: %d, but only %d were added successfully.", numberOfQuest, AddedQuestCount), http.StatusBadRequest, w)
		return
	}

	//Commit
	commitErr := tx.Commit()
	if commitErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to commit. %v", commitErr), http.StatusInternalServerError, w)
		return
	}

	writeResultJSON("Successfully saved mission", http.StatusOK, w)
}

func apiGetUserQuest(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)
	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
		return
	}

	//Get user id
	userID, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	//Get payload from body
	var payload map[string]interface{}

	DecodeErr := json.NewDecoder(r.Body).Decode(&payload)
	if DecodeErr != nil {
		writeResultJSON("Invalid JSON body", http.StatusBadRequest, w)
		return
	}

	missionType, bIsTypeValid := payload["type"].(string)

	if bIsTypeValid == false {
		writeResultJSON("type not found in payload", http.StatusBadRequest, w)
		return
	}

	//Get quest
	getUserQuestQuery := fmt.Sprintf(`
	SELECT umt.mission_id, umt.mission_name, umt.mission_progress,
	CONCAT('[',
		GROUP_CONCAT(
		CONCAT('{', 
			'"objective_id": "', mot.objective_id, '", ',
			'"objective_name": "', mot.objective_name, '", '
			'"objective_description": "', mot.objective_description, '", '
			'"objective_progress": ', mot.objective_progress
			,'}')
			ORDER BY mot.objective_order ASC
		)
	,']') AS objectives
	-- 
	FROM %s umt
	INNER JOIN (
	  SELECT mot1.*
	  FROM %s mot1
	  INNER JOIN (
	    SELECT objective_id, MAX(created_at) AS latest_created_at
	    FROM %s
	    GROUP BY objective_id
	  ) latest
	    ON mot1.objective_id = latest.objective_id AND mot1.created_at = latest.latest_created_at
	) mot
	  ON umt.mission_id = mot.mission_id
	WHERE umt.user_id = ? AND umt.mission_type = ?
	GROUP BY umt.mission_id
	`, dbTableMissionTracker, dbTableObjectiveTracker, dbTableObjectiveTracker)

	rows, queryErr := sqlDB.Query(getUserQuestQuery, userID, missionType)

	if queryErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to retrieve the mission data for the user. %v\n", queryErr), http.StatusBadRequest, w)
		return
	}

	defer rows.Close()

	type mission struct {
		MissionID       string                   `json:"mission_id"`
		MissionName     string                   `json:"mission_name"`
		MissionProgress string                   `json:"mission_progress"`
		Objectives      []map[string]interface{} `json:"objectives"`
	}

	var quests []mission

	for rows.Next() {
		var m mission
		var rawObjectives string
		scanErr := rows.Scan(&m.MissionID, &m.MissionName, &m.MissionProgress, &rawObjectives)

		if scanErr != nil {
			log.Printf("Failed to scan. %v \n", scanErr)
			continue
		}

		unmarshalObjectiveErr := json.Unmarshal([]byte(rawObjectives), &m.Objectives)

		log.Printf("Trying to unmarshal objective -> %s\n", m.MissionName)
		if unmarshalObjectiveErr != nil {
			log.Printf("Failed to unmarshal objective. %v \n", unmarshalObjectiveErr)
			continue
		}

		quests = append(quests, m)
	}

	jsonQuest, marshalQuestErr := json.Marshal(quests)

	if marshalQuestErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to retrieve the mission data for the user. %v\n", marshalQuestErr), http.StatusInternalServerError, w)
		return
	}

	writeResultJSON(string(jsonQuest), http.StatusOK, w)
}

func apiUpdateUserQuest(w http.ResponseWriter, r *http.Request) {

	token, tokenErr := checkIsTokenValid(w, r)
	if tokenErr != nil {
		writeResultJSON(fmt.Sprintf("%v", tokenErr), http.StatusUnauthorized, w)
		return
	}

	//Get user id
	_, userIDErr := getUserIDFromToken(token)

	if userIDErr != nil {
		writeResultJSON(fmt.Sprintf("%v", userIDErr), http.StatusNotFound, w)
		return
	}

	if sqlDB == nil {
		writeResultJSON("❌ sqlDB is null -> database connection not initialized.", http.StatusInternalServerError, w)
		return
	}

	//Get payload from body
	var payload map[string]interface{}

	DecodeErr := json.NewDecoder(r.Body).Decode(&payload)
	if DecodeErr != nil {
		writeResultJSON("Invalid JSON body", http.StatusBadRequest, w)
		return
	}

	objectiveID, bIsObjectiveIDValid := payload["objective_id"].(string)
	objectiveName, bIsObjectiveNameValid := payload["objective_name"].(string)
	objectiveDes, bIsObjectiveDesValid := payload["objective_description"].(string)
	objectiveProgress, bIsObjectiveProgressValid := payload["objective_progress"].(float64)

	if bIsObjectiveIDValid == false {
		writeResultJSON("objective_id not found in payload", http.StatusBadRequest, w)
		return
	}

	if bIsObjectiveNameValid == false {
		writeResultJSON("objective_name not found in payload", http.StatusBadRequest, w)
		return
	}

	if bIsObjectiveDesValid == false {
		writeResultJSON("objective_description not found in payload", http.StatusBadRequest, w)
		return
	}

	if bIsObjectiveProgressValid == false {
		writeResultJSON("objective_progress not found in payload", http.StatusBadRequest, w)
		return
	}

	tx, txErr := sqlDB.Begin()

	if txErr != nil {
		writeResultJSON("❌ Failed to start a transaction.", http.StatusInternalServerError, w)
		return
	}

	defer tx.Rollback()

	getObjectiveQuery := fmt.Sprintf(`SELECT mission_id, objective_order FROM %s WHERE objective_id = ?`, dbTableObjectiveTracker)

	var missionID string
	var objectiveOrder string
	scanErr := tx.QueryRow(getObjectiveQuery, objectiveID).Scan(&missionID, &objectiveOrder)

	if scanErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to query. %v", scanErr), http.StatusInternalServerError, w)
		return
	}

	InsertObjectiveQuery := fmt.Sprintf(`INSERT INTO %s (mission_id, objective_id, objective_order, objective_name, objective_description, objective_progress) VALUES ( ?, ?, ?, ?, ?, ?)`, dbTableObjectiveTracker)

	_, InsertObjectiveErr := tx.Exec(InsertObjectiveQuery, missionID, objectiveID, objectiveOrder, objectiveName, objectiveDes, objectiveProgress)

	if InsertObjectiveErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to update mission. %v", scanErr), http.StatusInternalServerError, w)
		return
	}

	// Update mission status based on its objectives
	// Get total number of objectives
	getTotalObjectiveQuery := fmt.Sprintf(`
	SELECT COUNT(*) AS total_Objective
	FROM %s mot
	WHERE
	mot.mission_id = ?
	-- 
	AND
	-- 
	mot.created_at = ( 
		SELECT MAX(mot2.created_at)
		FROM %s mot2
		WHERE mot.objective_id = mot2.objective_id
	)
	`, dbTableObjectiveTracker, dbTableObjectiveTracker)

	var TotalObjectiveCount int

	scanTotalObjectiveErr := tx.QueryRow(getTotalObjectiveQuery, missionID).Scan(&TotalObjectiveCount)

	if scanTotalObjectiveErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to get total number of objectives. %v", scanTotalObjectiveErr), http.StatusInternalServerError, w)
		return
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////

	//Update mission progress

	if TotalObjectiveCount == 0 {
		writeResultJSON("Failed to update mission because this mission has no objectives.", http.StatusInternalServerError, w)
		return
	}

	getTotalProgressQuery := fmt.Sprintf(`
	SELECT SUM(mot.objective_progress) AS TotalProgress
	FROM %s mot
	WHERE mot.mission_id = ?
	AND mot.created_at = (
	SELECT MAX(mot2.created_at) FROM %s mot2
	WHERE mot.objective_id = mot2.objective_id
	)`, dbTableObjectiveTracker, dbTableObjectiveTracker)

	var ObjectivesProgress int

	scanProgressErr := tx.QueryRow(getTotalProgressQuery, missionID).Scan(&ObjectivesProgress)

	if scanProgressErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to scan mission progress. %v", scanProgressErr), http.StatusInternalServerError, w)
		return
	}

	var missionProgress int
	missionProgress = (ObjectivesProgress / (TotalObjectiveCount * 100)) * 100

	//Insert a new record for the mission

	updateMissionQuery := fmt.Sprintf(`
	UPDATE %s 
	SET mission_progress = ?, 
	updated_at = CURRENT_TIMESTAMP
	WHERE mission_id = ?
	`, dbTableMissionTracker)

	_, updateMissionErr := tx.Exec(updateMissionQuery, missionProgress, missionID)

	if updateMissionErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to update mission. %v", updateMissionErr), http.StatusInternalServerError, w)
		return
	}

	//Commit
	commitErr := tx.Commit()
	if commitErr != nil {
		writeResultJSON(fmt.Sprintf("Failed to commit. %v", commitErr), http.StatusInternalServerError, w)
		return
	}

	writeResultJSON("Mission and objective updated successfully", http.StatusOK, w)
}
