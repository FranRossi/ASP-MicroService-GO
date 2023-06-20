package main

import (
	"net/http"
	"os"
	"user-service/cmd/routes"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	router := gin.Default()

	log.Info().Msg("Starting server...")

	//routes
	routes.UserRoute(router)
	port := os.Getenv("PORT")
	if port == "" {
		log.Info().Msg("No PORT environment variable detected, defaulting to 6000")
		port = "6000" // Default port for local development
	}

	// Start the server on the specified port
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Error().Err(err).Msg("Error starting server on port " + port)
		panic(err)
	}
}
