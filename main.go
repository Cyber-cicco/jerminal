package main

import . "github.com/Cyber-cicco/jerminal/jerminal"

// This is what i want my jerminal application workflow to look like
func main() {

	agent := GetAgent("agent1")

	// Doit être une fonction asynchrone
    pipeline := SetPipeline(
		Agent(agent),
		Stages(func (){
            Stage(),
        }),
    ),
}
