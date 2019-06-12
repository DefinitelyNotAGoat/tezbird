package gotezos

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// NodeService is a service for node related functions
type NodeService struct {
	gt *GoTezos
}

// Bootstrap is a structure representing the bootstrapped response
type Bootstrap struct {
	Block     string
	Timestamp time.Time
}

// newNodeService returns a new NodeService
func (gt *GoTezos) newNodeService() *NodeService {
	return &NodeService{gt: gt}
}

// Bootstrapped gets the current node bootstrap
func (n *NodeService) Bootstrapped() (Bootstrap, error) {
	var b Bootstrap
	query := "/monitor/bootstrapped"
	resp, err := n.gt.Get(query, nil)
	if err != nil {
		return b, errors.Wrapf(err, "could not node bootstraped '%s'", query)
	}

	b, err = unmarshallBootstrap(resp)
	if err != nil {
		return b, errors.Wrapf(err, "could not node bootstraped '%s'", query)
	}

	return b, nil
}

// CommitHash gets the current commit the node is running
func (n *NodeService) CommitHash() (string, error) {
	var c string
	query := "/monitor/commit_hash"
	resp, err := n.gt.Get(query, nil)
	if err != nil {
		return c, errors.Wrapf(err, "could not node commit hash '%s'", query)
	}

	c, err = unmarshalString(resp)
	if err != nil {
		return c, errors.Wrapf(err, "could not node commit hash '%s'", query)
	}

	return c, nil
}

func unmarshallBootstrap(v []byte) (Bootstrap, error) {
	b := Bootstrap{}
	err := json.Unmarshal(v, &b)
	if err != nil {
		return b, errors.Wrap(err, "could not unmarshal bytes into Bootstrap")
	}

	return b, nil
}
