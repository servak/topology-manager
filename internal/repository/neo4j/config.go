package neo4j

import "fmt"

// Neo4jConfig holds Neo4j database configuration
type Neo4jConfig struct {
	URI      string `yaml:"uri"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// DSN returns the Neo4j connection URI
func (c *Neo4jConfig) DSN() string {
	return c.URI
}

// Validate checks if the configuration is valid
func (c *Neo4jConfig) Validate() error {
	if c.URI == "" {
		return fmt.Errorf("neo4j URI is required")
	}
	if c.Username == "" {
		return fmt.Errorf("neo4j username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("neo4j password is required")
	}
	return nil
}