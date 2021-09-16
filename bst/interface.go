package bst

// Storage is general in-memory BST storage implementation
type Storage interface {
	// Insert inserts to a vertex with topic ident connection uuid
	Insert(uuid string, topic string)
	// Remove removes uuid from topic, if the uuid is single for a topic, whole vertex will be removed
	Remove(uuid, topic string)
	// Get will return all connections associated with the topic
	Get(topic string) map[string]struct{}
	// Contains checks if the BST contains a topic
	Contains(topic string) bool
}
