package game

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByOwnerName []*Room

func (a ByOwnerName) Len() int      { return len(a) }
func (a ByOwnerName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByOwnerName) Less(i, j int) bool {
	if a[i].owner == nil {
		return true
	}

	if a[j].owner == nil {
		return false
	}
	return a[i].owner.Name() < a[j].owner.Name()
}

type ByNumPlayers []*Room

func (a ByNumPlayers) Len() int      { return len(a) }
func (a ByNumPlayers) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNumPlayers) Less(i, j int) bool {
	if a[i].maxNumberOfPlayers-len(a[i].players.coreMap) == 0 {
		return true
	}
	if a[j].maxNumberOfPlayers-len(a[j].players.coreMap) == 0 {
		return false
	}
	return a[i].maxNumberOfPlayers-len(a[i].players.coreMap) > a[j].maxNumberOfPlayers-len(a[j].players.coreMap)
}

type ByRequirement []*Room

func (a ByRequirement) Len() int           { return len(a) }
func (a ByRequirement) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRequirement) Less(i, j int) bool { return a[i].requirement < a[j].requirement }
