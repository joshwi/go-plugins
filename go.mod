module github.com/joshwi/go-plugins

go 1.16

replace github.com/joshwi/go-plugins/graphdb => ./graphdb

require (
	github.com/joshwi/go-utils/parser v0.0.0-20211104230733-a6f685775666
	github.com/joshwi/go-utils/utils v0.0.0-20211104231639-7b678230ea04
	github.com/neo4j/neo4j-go-driver v1.8.3
)
