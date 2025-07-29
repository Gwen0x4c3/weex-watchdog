package weex

import (
	"sync"
)

// ContractMapper 合约映射管理器
type ContractMapper struct {
	contractMap map[string]string // contractId -> symbol name
	mu          sync.RWMutex
}

var (
	// globalContractMapper 全局合约映射实例
	globalContractMapper *ContractMapper
	once                 sync.Once
)

// GetContractMapper 获取全局合约映射实例
func GetContractMapper() *ContractMapper {
	once.Do(func() {
		globalContractMapper = &ContractMapper{
			contractMap: make(map[string]string),
		}
	})
	return globalContractMapper
}

// LoadContractMapping 加载合约映射
func (cm *ContractMapper) LoadContractMapping() error {
	contractMap, err := GetMetaDataV2()
	if err != nil {
		return err
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.contractMap = contractMap

	return nil
}

// GetSymbolName 根据合约ID获取交易对名称
func (cm *ContractMapper) GetSymbolName(contractID string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if symbolName, exists := cm.contractMap[contractID]; exists {
		return symbolName
	}

	// 如果找不到映射，返回原始的contractID
	return contractID
}

// GetAllMappings 获取所有映射（用于调试）
func (cm *ContractMapper) GetAllMappings() map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range cm.contractMap {
		result[k] = v
	}
	return result
}

// GetMappingCount 获取映射数量
func (cm *ContractMapper) GetMappingCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.contractMap)
}
