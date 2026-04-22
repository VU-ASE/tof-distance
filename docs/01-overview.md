import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Overview

## Purpose

The `distance` service reads measurements from a VL6180X Time-of-Flight distance sensor over I2C and publishes them as distance readings for other Rover services to use.

At the moment, the implementation publishes a single distance stream (`distance-one`). The code already contains placeholders for a second sensor and stream, but that part is currently disabled.

## Installation

To install this service, the latest release of [`roverctl`](https://ase.vu.nl/docs/framework/Software/rover/roverctl/installation) should be installed for your system and your Rover should be powered on.

<Tabs groupId="installation-method">
<TabItem value="roverctl" label="Using roverctl" default>

1. Install the service from your terminal

```bash
# Replace ROVER_NUMBER with your the number label on your Rover (e.g. 7)
roverctl service install -r <ROVER_NUMBER> https://github.com/VU-ASE/<SERVICE_REPOSITORY>/releases/latest/download/<SERVICE_NAME>.zip
```

</TabItem>
<TabItem value="roverctl-web" label="Using roverctl-web">

1. Open `roverctl-web` for your Rover

```bash
# Replace ROVER_NUMBER with your the number label on your Rover (e.g. 7)
roverctl -r <ROVER_NUMBER>
```

2. Click on "install a service" button on the bottom left, and click "install from URL"
3. Enter the URL of the latest release:

```text
https://github.com/VU-ASE/<SERVICE_REPOSITORY>/releases/latest/download/<SERVICE_NAME>.zip
```

</TabItem>
</Tabs>

Follow [this tutorial](https://ase.vu.nl/docs/tutorials/write-a-service/upload) to understand how to use an ASE service. You can find more useful `roverctl` commands [here](/docs/framework/Software/rover/roverctl/usage)

## Requirements

- A VL6180X Time-of-Flight distance sensor should be connected over I2C
- A TCA9548A-compatible I2C multiplexer should be connected if the sensor is accessed through multiplexed channels
- The correct I2C bus and channel values should be configured in this service's `service.yaml`

## Inputs

As defined in the `service.yaml`, this service does not depend on any read streams.

## Outputs

As defined in the `service.yaml`, this service exposes the following write streams:

- `distance-one`:
  - To this stream, [`DistanceSensorOutput`](https://github.com/VU-ASE/rovercom/blob/main/definitions/outputs/distance.proto) messages will be written, wrapped in a [`SensorOutput` wrapper message](https://github.com/VU-ASE/rovercom/blob/main/definitions/outputs/wrapper.proto)

The current implementation only writes to `distance-one`. Support for a second output stream is present in the codebase, but is commented out at the moment.