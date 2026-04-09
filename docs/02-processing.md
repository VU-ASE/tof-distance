# Processing

This service follows the following steps:

1. It reads its configuration values from the `service.yaml` file:
   - `bus`: the I2C bus number used to communicate with the sensor and multiplexer
   - `channel-one`: the multiplexer channel on which the sensor is connected
   - `frame-rate`: the rate at which new distance measurements should be published

2. It initializes a VL6180X Time-of-Flight sensor over I2C. During initialization, the service:
   - opens the sensor's I2C address
   - opens the I2C multiplexer address
   - selects the configured channel
   - checks whether the sensor is in a fresh-reset state
   - writes the mandatory startup registers if initialization is needed

3. In its main loop, the service triggers a ranging measurement on the sensor and waits until the measurement is ready

4. Once a valid reading is available, it reads the measured distance value from the sensor. The raw value is read in millimeters and converted to meters before publishing

5. The distance value is encoded in a [`DistanceSensorOutput`](https://github.com/VU-ASE/rovercom/blob/main/definitions/outputs/distance.proto) message, which is then wrapped in a [`SensorOutput`](https://github.com/VU-ASE/rovercom/blob/main/definitions/outputs/wrapper.proto) message

6. The wrapped message is written to the `distance-one` stream

7. Finally, the service waits according to the configured `frame-rate` and repeats the process

## Notes

- The current implementation publishes only one distance sensor stream (`distance-one`)
- The codebase contains commented-out scaffolding for a second sensor (`distance-two`), but this is not currently active
- If `frame-rate` is set to `0`, the service returns an error to avoid division by zero
- If a sensor read fails, the service logs the error and continues trying to read again

## About the VL6180X Sensor

The VL6180X is a Time-of-Flight (ToF) distance sensor. Unlike simple infrared proximity sensors, it estimates distance by measuring the time it takes for emitted light to reflect off an object and return to the sensor. This makes it more suitable for short-range distance measurements where a direct metric reading is needed.

In this service, the sensor is used to produce real-time distance measurements that can be consumed by other Rover services for obstacle awareness, navigation logic, or higher-level decision making.