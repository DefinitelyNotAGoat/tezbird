package gotezos

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// NetworkService is wrapper representing network functions
type NetworkService struct {
	gt *GoTezos
}

// NetworkVersion represents the network version returned by the Tezos network.
type NetworkVersion struct {
	Name    string `json:"name"`
	Major   int    `json:"major"`
	Minor   int    `json:"minor"`
	Network string // Human readable network name
}

// NetworkVersions is an array of NetworkVersion
type NetworkVersions []NetworkVersion

// NetworkConstants represents the network constants returned by the Tezos network.
type NetworkConstants struct {
	ProofOfWorkNonceSize         int      `json:"proof_of_work_nonce_size"`
	NonceLength                  int      `json:"nonce_length"`
	MaxRevelationsPerBlock       int      `json:"max_revelations_per_block"`
	MaxOperationDataLength       int      `json:"max_operation_data_length"`
	MaxProposalsPerDelegate      int      `json:"max_proposals_per_delegate"`
	PreservedCycles              int      `json:"preserved_cycles"`
	BlocksPerCycle               int      `json:"blocks_per_cycle"`
	BlocksPerCommitment          int      `json:"blocks_per_commitment"`
	BlocksPerRollSnapshot        int      `json:"blocks_per_roll_snapshot"`
	BlocksPerVotingPeriod        int      `json:"blocks_per_voting_period"`
	TimeBetweenBlocks            []string `json:"time_between_blocks"`
	EndorsersPerBlock            int      `json:"endorsers_per_block"`
	HardGasLimitPerOperation     string   `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock         string   `json:"hard_gas_limit_per_block"`
	ProofOfWorkThreshold         string   `json:"proof_of_work_threshold"`
	TokensPerRoll                string   `json:"tokens_per_roll"`
	MichelsonMaximumTypeSize     int      `json:"michelson_maximum_type_size"`
	SeedNonceRevelationTip       string   `json:"seed_nonce_revelation_tip"`
	OriginationSize              int      `json:"origination_size"`
	BlockSecurityDeposit         string   `json:"block_security_deposit"`
	EndorsementSecurityDeposit   string   `json:"endorsement_security_deposit"`
	BlockReward                  string   `json:"block_reward"`
	EndorsementReward            string   `json:"endorsement_reward"`
	CostPerByte                  string   `json:"cost_per_byte"`
	HardStorageLimitPerOperation string   `json:"hard_storage_limit_per_operation"`
}

// Connections represents network connections
type Connections []struct {
	Incoming bool   `json:"incoming"`
	PeerID   string `json:"peer_id"`
	IDPoint  struct {
		Addr string `json:"addr"`
		Port int    `json:"port"`
	} `json:"id_point"`
	RemoteSocketPort int `json:"remote_socket_port"`
	Versions         []struct {
		Name  string `json:"name"`
		Major int    `json:"major"`
		Minor int    `json:"minor"`
	} `json:"versions"`
	Private       bool `json:"private"`
	LocalMetadata struct {
		DisableMempool bool `json:"disable_mempool"`
		PrivateNode    bool `json:"private_node"`
	} `json:"local_metadata"`
	RemoteMetadata struct {
		DisableMempool bool `json:"disable_mempool"`
		PrivateNode    bool `json:"private_node"`
	} `json:"remote_metadata"`
}

// NewNetworkService returns a new NetworkService
func (gt *GoTezos) newNetworkService() *NetworkService {
	return &NetworkService{gt: gt}
}

// GetVersions gets the network versions of Tezos network the client is using.
func (n *NetworkService) GetVersions() ([]NetworkVersion, error) {
	query := "/network/versions"
	networkVersions := make([]NetworkVersion, 0)
	resp, err := n.gt.Get(query, nil)
	if err != nil {
		return networkVersions, errors.Wrapf(err, "could not get network versions '%s'", query)
	}

	var nvs NetworkVersions
	nvs, err = nvs.unmarshalJSON(resp)
	if err != nil {
		return networkVersions, errors.Wrapf(err, "could not get network versions '%s'", query)
	}

	// Extract just the network name and append to returning slice
	// 'range' operates on a copy of the struct so cannot update-in-place
	for _, v := range nvs {

		parts := strings.Split(v.Name, "_")
		if len(parts) == 3 {
			v.Network = parts[1]
		}

		networkVersions = append(networkVersions, v)
	}

	return networkVersions, nil
}

// GetConstants gets the network constants for the Tezos network the client is using.
func (n *NetworkService) GetConstants() (NetworkConstants, error) {
	query := "/chains/main/blocks/head/context/constants"
	networkConstants := NetworkConstants{}
	resp, err := n.gt.Get(query, nil)
	if err != nil {
		return networkConstants, errors.Wrapf(err, "could not get network constants '%s'", query)
	}
	networkConstants, err = networkConstants.unmarshalJSON(resp)
	if err != nil {
		return networkConstants, errors.Wrapf(err, "could not get network constants '%s'", query)
	}

	return networkConstants, nil
}

// GetChainID gets the id of the chain with the most fitness
func (n *NetworkService) GetChainID() (string, error) {
	query := "/chains/main/chain_id"
	resp, err := n.gt.Get(query, nil)
	if err != nil {
		return "", errors.Wrapf(err, "could not get chain ID '%s'", query)
	}

	chainID, err := unmarshalString(resp)
	if err != nil {
		return "", errors.Wrapf(err, "could not get chain ID '%s'", query)
	}

	return chainID, nil
}

// Connections gets the network connections
func (n *NetworkService) Connections() (Connections, error) {
	var connections Connections
	query := "/network/connections"
	resp, err := n.gt.Get(query, nil)
	if err != nil {
		return connections, errors.Wrapf(err, "could not get network connections '%s'", query)
	}

	connections, err = connections.unmarshalConnections(resp)
	if err != nil {
		return connections, errors.Wrapf(err, "could not get network connections '%s'", query)
	}

	return connections, nil
}

// UnmarshalJSON unmarshals the bytes received as a parameter, into the type NetworkVersion.
func (nvs *NetworkVersions) unmarshalJSON(v []byte) (NetworkVersions, error) {
	networkVersions := NetworkVersions{}
	err := json.Unmarshal(v, &networkVersions)
	if err != nil {
		return networkVersions, errors.Wrap(err, "could not unmarshal bytes into NetworkVersions")
	}
	return networkVersions, nil
}

// unmarshalConnections unmarshals the bytes received as a parameter, into the type Connections.
func (c *Connections) unmarshalConnections(v []byte) (Connections, error) {
	connections := Connections{}
	err := json.Unmarshal(v, &connections)
	if err != nil {
		return connections, errors.Wrap(err, "could not unmarshal bytes into Connections")
	}
	return connections, nil
}

// UnmarshalJSON unmarshals bytes received as a parameter, into the type NetworkConstants.
func (nc *NetworkConstants) unmarshalJSON(v []byte) (NetworkConstants, error) {
	networkConstants := NetworkConstants{}
	err := json.Unmarshal(v, &networkConstants)
	if err != nil {
		return networkConstants, errors.Wrap(err, "could not unmarshal bytes into NetworkConstants")
	}
	return networkConstants, nil
}
