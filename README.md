# shardstore

Thread safe store library in Go that keeps a set of IDs in a very specific way. The Store can be used as shared resource by multiple system components. It is thread safe library. IDs can be any arbitrary strings.

## Library Interface

- `ReportIDs(ID string, shard int) []string, []string` - Store will collect IDs for a provided shard, it will determine a common set with the existing shard list returning back the difference: new items and removals. The set within the same shard will be deduplicated.

- `All() []string` will dynamically calculate and return a deterministically ordered list of unique IDs currently present in the entire Store from all shards.

- `Diff(map[string]int newMapping) map[string]MetaInfo, []string` based on the provided mapping it will dynamically return the information about current vs. new mapping. It will return detailed MetaInfo structure `MetaInfo{oldShard:0 int, newShard: int, changed: bool} and a list of WorkflowIDs that are subject to change the shard.

- `Duplicates() []string []int` it will return a list of IDs that are currently assigned to more then 1 shard. This is a check function for validation - the correct implementation has state guarantees that it should never present duplicates.

## Licence

