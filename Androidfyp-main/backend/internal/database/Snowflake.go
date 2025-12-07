package db

import (
	"log"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func InitIDGenerator() {
	var err error
	node, err = snowflake.NewNode(1)
	if err != nil {
		log.Fatal(err)
	}
}

// Exported function to generate IDs
func GenerateID() uint64 {
	return uint64(node.Generate().Int64())
}
