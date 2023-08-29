package main

import "order_notification_system/cmd/rest"

func main() {
	err := rest.ServeRest()
	if err != nil {
		panic(err)
	}
}
