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
	// distanceStreamTwo := service.GetWriteStream("distance-two")

	channelOne, err := configuration.GetFloatSafe("channel-one")
	if err != nil {
		return fmt.Errorf("Failed to get configuration: %v", err)
	}

	// channelTwo, err = configuration.GetFloatSafe("channel-two")
	// if err != nil {
	// 	return fmt.Errorf("Failed to get configuration: %v", err)
	// }
	
	
	channel2, err := configuration.GetFloatSafe("second-channel")
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
	sensor, err := Initialize(uint(math.Round(bus)), [2]uint8{uint8(math.Round(channelOne)), uint8(math.Round(channelOne))})
	if err != nil {
		log.Error().Msgf(err.Error())
		return fmt.Errorf("Failed to initialize Time Of Flight sensor")
	}

	for{
		distanceOne, err := sensor.ReadDistance(0)
		if err != nil {
			log.Error().Msgf("Error reading...%v", err)
			continue
		} 
		time.Sleep(time.Millisecond * 5)
		// distanceTwo, err := sensor.ReadDistance(1)
		// if err != nil {
		// 	log.Error().Msgf("Error reading sensor...%v", err)
		// 	continue
		// } 
		err = distanceStreamOne.Write(
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
		// errTwo := distanceStreamTwo.Write(
		// 	&pb_outputs.SensorOutput{
		// 		SensorId:  2,
		// 		Status:    0,
		// 		Timestamp: uint64(time.Now().UnixMilli()),
		// 		SensorOutput: &pb_outputs.SensorOutput_DistanceOutput{
		// 			DistanceOutput: &pb_outputs.DistanceSensorOutput{
		// 				Distance: float32(distanceTwo) / 1000.0,
		// 			},
		// 		},
		// 	},
		// )
		if err != nil{
			log.Err(err).Msg("Failed to send distance output")
			continue
		}

		log.Info().Msgf("distance sensor one: %f m", float32(distanceOne) / 1000.0)
		//log.Info().Msgf("distance sensor two: %f m", float32(distanceTwo) / 1000.0)

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
