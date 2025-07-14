package main

import (
	customhttp "weather-api-project/modules/customhttp"
	redis "weather-api-project/modules/redis"
)

func main() {

	redis.StartRedisServer()

	customhttp.StartHosting()
}
