package utils

// 定义挂载的磁盘
type Disk struct {
	Type string `json:"type"` // 磁盘类型
	Path string `json:"path"` // 磁盘路径
}

// 定义网卡
type Network struct {
	Type string `json:"type"` // 网卡类型
}

// 定义机器信息
type Machine struct {
	Name        string    `json:"name"`        // 系统名
	Memory      int64     `json:"memory"`      // 内存，单位M
	Cpu         int8      `json:"cpu"`         // cpu，单位核
	MultDisk    []Disk    `json:"multdisk"`    // 磁盘组
	MultNetwork []Network `json:"multnetwork"` // 网卡组
}

func (m *Machine) NewMachine() {

}

func (m *Machine) Craete() error {
	return nil
}
