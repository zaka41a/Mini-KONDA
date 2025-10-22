package store

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4j struct {
	driver neo4j.DriverWithContext
}

func NewNeo4j(uri, user, pass string) (*Neo4j, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, pass, ""))
	if err != nil {
		return nil, err
	}
	return &Neo4j{driver: driver}, nil
}

func (n *Neo4j) Close(ctx context.Context) { _ = n.driver.Close(ctx) }

func (n *Neo4j) SaveAnnotation(ctx context.Context, column, annotation string) error {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
	MERGE (c:Column {name: $column})
	MERGE (a:Annotation {text: $annotation})
	MERGE (c)-[:DESCRIBED_AS]->(a)
	`
	_, err := session.Run(ctx, query, map[string]any{
		"column":     column,
		"annotation": annotation,
	})
	if err != nil {
		return fmt.Errorf("neo4j run: %w", err)
	}
	return nil
}
