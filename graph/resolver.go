package graph

import (
	"social/internal/store/mongodb"
)

type Resolver struct {
	MongoStore mongodb.MongoStorage
}
