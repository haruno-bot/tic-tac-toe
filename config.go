package toe

// Config 井字棋插件配置信息
type Config struct {
	TicTacToe struct {
		Name      string  `toml:"name"`
		Version   string  `toml:"version"`
		GroupNums []int64 `toml:"groupNums"`
	} `toml:"tic-tac-toe"`
}
