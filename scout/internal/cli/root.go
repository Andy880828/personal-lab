package cli

// Root 回傳指令樹。describe 從同一棵樹產生 manifest，文件不會跟行為脫節。
func Root() *Command {
	return &Command{
		Name:    "scout",
		Summary: "agent-operated tech news radar: fetch your own sources, dedupe, rank transparently, let the agent pick and summarise",
		Sub: []*Command{
			statusCmd(),
			initCmd(),
			fetchCmd(),
			candidatesCmd(),
			readCmd(),
			{Name: "digest", Summary: "daily picks written by the agent", Sub: []*Command{
				digestCreateCmd(), digestShowCmd(), digestListCmd(), digestExportCmd(),
			}},
			trendsCmd(),
			{Name: "sources", Summary: "manage the watchlist sources", Sub: []*Command{
				sourcesListCmd(), sourcesAddCmd(), sourcesRmCmd(), sourcesToggleCmd(true), sourcesToggleCmd(false), sourcesTestCmd(),
			}},
			{Name: "keywords", Summary: "manage boost / mute keywords", Sub: []*Command{
				keywordsListCmd(), keywordsEditCmd(true), keywordsEditCmd(false),
			}},
			pruneCmd(),
			describeCmd(),
			versionCmd(),
		},
	}
}
