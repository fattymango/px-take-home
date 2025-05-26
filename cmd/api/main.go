package main

import (
	"fmt"

	"github.com/fattymango/px-take-home/app"
	"github.com/pkg/profile"
)

const (
	VERSION = "0.0.1"
)

// @title PX Take Home API Specification
// @version 0.0.1
// @description This is the API specification for the PX Take Home API
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8888
// @BasePath /
// @schemes http

// @securityDefinitions.basic BasicAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	defer profile.Start(profile.CPUProfile, profile.ProfilePath("./profile/cpu")).Stop()

	fmt.Println("Starting PX Take Home API Server version", VERSION)
	app.Start()
}
