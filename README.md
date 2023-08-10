## Overview

This project was part of an assignment in CS3211 Parallel and Concurrent Programming, taken in National University of Singapore.
The goal of this assignment was to design an exchange matching engine that would handle orders concurrent orders.

The design we settled on is shown as follows:

![image]("./program_design.jpeg")

The arrows represent channels and the circles represent goroutines. The capacity of the channels are noted with underlined numbers.
A more detailed explanation is written in the report

With this design, the engine can handle up to 1 buy and 1 sell order for each instrument at any point in time. 

