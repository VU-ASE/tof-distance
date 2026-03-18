package main

import (
	"fmt"
	"math"
	"os"
	"time"

	pb_outputs "github.com/VU-ASE/rovercom/v2/packages/go/outputs"
	roverlib "github.com/VU-ASE/roverlib-go/v2/src"

	"github.com/rs/zerolog/log"
)

func run(service roverlib.Service, configuration *roverlib.ServiceConfiguration) error {

	//
	// Acquire distance output streams
	//
	distanceStreamOne := service.GetWriteStream("distance-one")
	distanceStreamTwo := service.GetWriteStream("distance-two")

	//
	// Read multiplex configuration value
	//
	multiplex, err := configuration.GetFloatSafe("multiplex")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
	}

	//
	// We are only interested in the channel if we use a multiplexer
	//
	channelOne := 0.0
	channelTwo := 0.0
	if multiplex > 0.5 {
		channelOne, err = configuration.GetFloatSafe("channel-one")
		if err != nil {
			return fmt.Errorf("Failed to get configuration: %v", err)
		}

		channelTwo, err = configuration.GetFloatSafe("channel-two")
		if err != nil {
			return fmt.Errorf("Failed to get configuration: %v", err)
		}
	}
	
	//
	// Read bus configuration value
	//
	bus, err := configuration.GetFloatSafe("bus")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
	}

	//
	// Read fps configuration value
	//
	frameRate, err := configuration.GetFloatSafe("frame-rate")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
	}

	//
	// Initialize our Time Of Flight sensor
	//
	sensor, err := Initialize(multiplex > 0.5, uint(math.Round(bus)), [2]uint8{uint8(math.Round(channelOne)), uint8(math.Round(channelTwo))})
	if err != nil {
		log.Error().Msgf(err.Error())
		return fmt.Errorf("Failed to initialize Time Of Flight sensor")
	}

	for{
		distanceOne, err := sensor.ReadDistance(0)
		if err != nil {
			log.Error().Msgf("Error reading...")
			continue
		} 
		distanceTwo, err := sensor.ReadDistance(1)
		if err != nil {
			log.Error().Msgf("Error reading...")
			continue
		} 
		errOne := distanceStreamOne.Write(
			&pb_outputs.SensorOutput{
				SensorId:  2,
				Status:    0,
				Timestamp: uint64(time.Now().UnixMilli()),
				SensorOutput: &pb_outputs.SensorOutput_DistanceOutput{
					DistanceOutput: &pb_outputs.DistanceSensorOutput{
						Distance: float32(distanceOne) / 1000.0,
					},
				},
			},
		)
		errTwo := distanceStreamTwo.Write(
			&pb_outputs.SensorOutput{
				SensorId:  2,
				Status:    0,
				Timestamp: uint64(time.Now().UnixMilli()),
				SensorOutput: &pb_outputs.SensorOutput_DistanceOutput{
					DistanceOutput: &pb_outputs.DistanceSensorOutput{
						Distance: float32(distanceTwo) / 1000.0,
					},
				},
			},
		)
		if errOne != nil || errTwo != nil {
			log.Err(err).Msg("Failed to send distance output")
			continue
		}

		log.Info().Msgf("distance sensor one: %f m", float32(distanceOne) / 1000.0)
		log.Info().Msgf("distance sensor two: %f m", float32(distanceTwo) / 1000.0)

		frameRate, err = configuration.GetFloat("frame-rate")
		if err != nil {
			return fmt.Errorf("Failed to get configuration: %v", err)
		}
		if frameRate == 0 {
			return fmt.Errorf("Frame rate can't be 0 (division by zero)")
		}

		time.Sleep(time.Second / time.Duration(frameRate))
		
	}
	
}

// This function gets called when roverd wants to terminate the service
func onTerminate(sig os.Signal) error {
	log.Info().Str("signal", sig.String()).Msg("Terminating service")
	return nil
}

// This is just a wrapper to run the user program
// it is not recommended to put any other logic here
func main() {
	roverlib.Run(run, onTerminate)
}
