package actor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopActor(t *testing.T) {
	const testPattern = "test-stopping-pattern"
	// Setup
	service := setup(testPattern)

	// Register a test pattern
	RegFactory(testPattern, func(actorName string) Behavior {
		return &StoppingBehavior{
			actorName: actorName,
		}
	})

	// Create our test actor
	actorName := "test-actor-stopping"
	actorRef := NewActorRef(NewProps(), actorName, testPattern)
	actorRef.Send("initial-message")
	// Stop the actor which will trigger the stopping phase
	err := StopActor(actorName, testPattern)
	assert.NoError(t, err)
	time.Sleep(time.Second * 10)
	service.Stop()
}

// TestMessageLossDuringStopping demonstrates that messages sent to an actor during its stopping phase may be lost
func TestMessageLossDuringStopping(t *testing.T) {
	const testPattern = "test-stopping-pattern"
	// Setup
	service := setup(testPattern)

	// Create channels to track message reception
	messageSent := make(chan bool, 1)

	// Register a test pattern
	RegFactory(testPattern, func(actorName string) Behavior {
		return &StoppingBehavior{
			actorName: actorName,
		}
	})

	// Create our test actor
	actorName := "test-actor-stopping"
	actorRef := NewActorRef(NewProps(), actorName, testPattern)
	actorRef.Send("initial-message")

	// Start a goroutine that will send messages during the stopping phase
	go func() {
		// Wait a bit before sending the message to allow the stop process to begin
		time.Sleep(10 * time.Millisecond)

		// Send a message while the actor is stopping
		res, err := actorRef.RequestFuture("stopping-phase-message")
		if err != nil {
			panic(err)
		}
		fmt.Printf("receive response: %v\n", res)
		messageSent <- true
	}()

	// Stop the actor which will trigger the stopping phase
	err := StopActor(actorName, testPattern)
	assert.NoError(t, err)

	// Wait for the message to be sent
	<-messageSent

	service.Stop()
}
