package datacenter

type DataObject interface {
	Id() int64
	SetId(id int64)
	CacheKey() string
	DatabaseTableName() string
	ClassName() string
}
