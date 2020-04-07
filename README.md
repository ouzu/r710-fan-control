[![Build Status](https://ci.laze.today/api/badges/ouzu/r710-fan/status.svg)](https://ci.laze.today/ouzu/r710-fan)

# R710 fan control

This little program controls the fans of my Dell R710 server.

It is inspired by [this program](https://github.com/sulaweyo/r710-fan-control) (big thanks to sulaweyo and NoLooseEnds) and designed with the following goals in mind:

- The server should run as silent as possible

- There should be no sudden changes in fan speed as I find it annoying

- It shouldn't damage my hardware

## Background

The Dell R710 servers have an iDRAC controller built in. This is basically a management interface and even has a separate Ethernet Port.

The fan control happens in this chip and is a black box and the fan speeds of this controller seem way too high at lower temperatures.

As my server lives right next to my desk and in the same room as my bed, I had to do something about it. I looked at existing solutions but weren't satisfied with the results.

## Problems

### Sudden speed changes

Existing programs all used some kind of temperature zones and assigned different fan speeds for different temperature ranges. With this approach my server showed the following behavior:

1. Temperature in a lower zone, fan speed low
2. Temperature rises
3. Temperature enters the next zone, fan speed increased
4. Temperature falls
5. Repeat from step 1

This cycle would repeat every few minutes and annoyed me.

### Bad sensors

During my experiments, I noticed that the temperature sensors seem to be really bad and can jump around a few degrees even when the server is idle.

## Solutions

The fan speed is set based on a function created using polynomial interpolation.

I chose a few values which seemed good and looked where the temperature would settle under different loads.

The function I use has the following values: 0% at 25°C, 10% at 45°C, 100% at 80°C.

To mitigate the sensor issues, an average of the last 4 reads is used as long as the temperatures don't increase by more than 10%.

To not damage my hardware, the fan control is handed back to the system controller if the temperature is unusually high.

## Results

Under average load the highest core temperature stays around 45°C which is fine.

When the load increases, the fan speed gradually increases and makes no sudden jumps :)

## Usage

The usage is a bit clumsy, see `./r710-fan -help`

A few examples:

- Enable the speed control: `./r710-fan -mode auto`

- Print the highest core temperature: `./r710-fan -mode print`

- Turn the server into a jet engine: `./r710-fan -mode manual -speed 100`

- Give back the control: `./r710-fan reset`