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
	// Acquire distance output stream
	//
	distanceStream1 := service.GetWriteStream("distance-1")
	distanceStream2 := service.GetWriteStream("distance-2")

	//
	// Read both channel configuration values
	//
	channel1, err := configuration.GetFloatSafe("channel-1")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
	}
	
	channel2, err := configuration.GetFloatSafe("channel-2")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
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
	// Initialize our Time Of Flight sensors
	//
	sensor1, err := Initialize(uint(math.Round(bus)), uint8(math.Round(channel1)))
	if err != nil {
		log.Error().Msgf(err.Error())
		return fmt.Errorf("Failed to initialize Time Of Flight sensor")
	}

	sensor2, err := Initialize(uint(math.Round(bus)), uint8(math.Round(channel2)))
	if err != nil {
		log.Error().Msgf(err.Error())
		return fmt.Errorf("Failed to initialize Time Of Flight sensor")
	}

	for{
		//
		// Read sensor 1
		//
		distance1, err := sensor1.ReadDistance()
		if err != nil {
			log.Error().Msgf("Sensor 1 failed to read distance: %v", err)
		} 

		err = distanceStream1.Write(
			&pb_outputs.SensorOutput{
				SensorId:  1,
				Status:    0,
				Timestamp: uint64(time.Now().UnixMilli()),
				SensorOutput: &pb_outputs.SensorOutput_DistanceOutput{
					DistanceOutput: &pb_outputs.DistanceSensorOutput{
						Distance: float32(distance1) / 1000.0,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("Failed to send distance output: %v", err)
		}

		log.Info().Msgf("Sensor 1 distance: %f m", float32(distance1) / 1000.0)

		//
		// Read sensor 2
		//
		distance2, err := sensor2.ReadDistance()
		if err != nil {
			log.Error().Msgf("Sensor 2 failed to read distance: %v", err)
		}

		err = distanceStream2.Write(
			&pb_outputs.SensorOutput{
				SensorId:  2,
				Status:    0,
				Timestamp: uint64(time.Now().UnixMilli()),
				SensorOutput: &pb_outputs.SensorOutput_DistanceOutput{
					DistanceOutput: &pb_outputs.DistanceSensorOutput{
						Distance: float32(distance2) / 1000.0,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("Failed to send distance output: %v", err)
		}

		log.Info().Msgf("Sensor 2 distance: %f m", float32(distance2) / 1000.0)


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

func onTerminate(sig os.Signal) error {
	log.Info().Str("signal", sig.String()).Msg("Terminating service")
	return nil
}

func main() {
	roverlib.Run(run, onTerminate)
}
