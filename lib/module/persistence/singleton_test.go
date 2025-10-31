package persistence

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSingleton(t *testing.T) {
	InitSingletonPersistenceWithFile("mongodb.toml")

	assert.NotNil(t, globalPersistence)

	id := "1234567890"
	resp, err := PersistSync(context.Background(), "test", "player", id, bson.M{"$set": bson.M{"name": "test", "age": 20}})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	fmt.Printf("ModifiedCount: %d, MatchedCount: %d\n", resp.ModifiedCount, resp.MatchedCount)
	err = SingletonGracefulStop()
	assert.NoError(t, err)
}
