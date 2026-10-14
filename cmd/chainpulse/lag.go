package main

type RPCStatus string

const (
	StatusHealthy RPCStatus = "HEALTHY"
	StatusLagging RPCStatus = "LAGGING"
)

type LagResult struct {
	RPCName     string
	ChainID     uint64
	BlockNumber uint64
	Lag         uint64
	Status      RPCStatus
}

func highestBlockByChain(results []*RPCResult) map[uint64]uint64 {
	highest := make(map[uint64]uint64)

	for _, result := range results {
		if result.Success == nil {
			continue
		}
		chainID := result.Success.ChainID
		blockNumber := result.Success.BlockNumber

		current, exists := highest[chainID]
		if !exists || blockNumber > current {
			highest[chainID] = blockNumber
		}
	}

	return highest
}

func calculateLag(results []*RPCResult, highest map[uint64]uint64) map[string]uint64 {

	lag := make(map[string]uint64)

	for _, result := range results {
		if result.Success == nil {
			continue
		}
		chainID := result.Success.ChainID
		blockNumber := result.Success.BlockNumber

		lag[result.Success.RPCName] = highest[chainID] - blockNumber
	}
	return lag
}

func determineStatus(lag uint64, threshold uint64) RPCStatus {

	if lag <= threshold {
		return StatusHealthy
	}
	return StatusLagging
}

func analyzeLag(results []*RPCResult, threshold uint64) []LagResult {
	highest := highestBlockByChain(results)
	lag := calculateLag(results, highest)

	var lagResults []LagResult
	for _, result := range results {
		if result.Success == nil {
			continue
		}
		success := result.Success
		rpcLag := lag[result.Success.RPCName]
		lagResult := LagResult{
			ChainID:     success.ChainID,
			RPCName:     success.RPCName,
			BlockNumber: success.BlockNumber,
			Lag:         rpcLag,
			Status:      determineStatus(rpcLag, threshold),
		}
		lagResults = append(lagResults, lagResult)
	}

	return lagResults
}

func hasUnhealthyRPCs(results []*RPCResult, threshold uint64) bool {
	for _, result := range results {
		if result.Failed != nil {
			return true
		}
	}

	resultLags := analyzeLag(results, threshold)

	for _, resultLag := range resultLags {
		if resultLag.Status == StatusLagging {
			return true
		}
	}

	return false
}
