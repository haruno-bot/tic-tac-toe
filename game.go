package toe

import (
	"fmt"
	"strings"

	"github.com/haruno-bot/haruno/logger"

	"github.com/BurntSushi/toml"

	"github.com/haruno-bot/haruno/coolq"
)

var name string
var version string

var groupNums = make(map[int64]bool)
var gameStarted = make(map[int64]bool)
var gameBoards = make(map[int64][][]int)
var gameWeight = make(map[int64][][]int)

const cX = "X"
const cO = "O"

var cMap = map[int]string{
	0: "-",
	1: cX,
	2: cO,
}
var validPosInput = map[string]bool{
	"A1": true, "A2": true, "A3": true,
	"B1": true, "B2": true, "B3": true,
	"C1": true, "C2": true, "C3": true,
}

// Game 井字棋游戏插件
type Game struct {
	coolq.Plugin
}

// Name 插件名字+版本号
func (_plugin Game) Name() string {
	return fmt.Sprintf("%s@%s", name, version)
}

func (_plugin *Game) loadConfig() error {
	cfg := new(Config)
	toml.DecodeFile("cofig.toml", cfg)
	_, err := toml.DecodeFile("config.toml", cfg)
	if err != nil {
		return err
	}
	pcfg := cfg.TicTacToe
	name = pcfg.Name
	version = pcfg.Version
	for _, groupID := range pcfg.GroupNums {
		groupNums[groupID] = true
		gameStarted[groupID] = false
		resetGameBoard(groupID)
	}
	return nil
}

func resetGameBoard(groupID int64) {
	gameBoards[groupID] = make([][]int, 3)
	gameWeight[groupID] = make([][]int, 3)
	for i := 0; i < 3; i++ {
		gameBoards[groupID][i] = make([]int, 3)
		gameWeight[groupID][i] = make([]int, 3)
		for j := 0; j < 3; j++ {
			gameBoards[groupID][i][j] = 0
			gameWeight[groupID][i][j] = 1
		}
	}
}

func displayGameBoard(groupID int64, reply coolq.Message) coolq.Message {
	reply = append(reply, coolq.NewTextSection("\tA\tB\tC"))
	for i, ln := range gameBoards[groupID] {
		reply = append(reply, coolq.NewTextSection(fmt.Sprintf("%d\t%s\t%s\t%s", i+1, cMap[ln[0]], cMap[ln[1]], cMap[ln[2]])))
	}
	return reply
}

// Filters 过滤酷Q上报事件用，利于提升插件性能
func (_plugin Game) Filters() map[string]coolq.Filter {
	filters := make(map[string]coolq.Filter)
	filters["tic-tac-toe-game-start"] = func(event *coolq.CQEvent) bool {
		if event.PostType != "message" ||
			event.MessageType != "group" ||
			event.SubType != "normal" {
			return false
		}
		if !groupNums[event.GroupID] {
			return false
		}
		msg := new(coolq.Message)
		err := coolq.Unmarshal([]byte(event.Message), msg)
		if err != nil {
			logger.Field(_plugin.Name()).Error(err.Error())
			return false
		}
		sec := (*msg)[0]
		if sec.Type == "text" && sec.Data["text"] == "# 井字棋" {
			return true
		}
		return false
	}
	filters["tic-tac-toe-gaming"] = func(event *coolq.CQEvent) bool {
		if event.PostType != "message" ||
			event.MessageType != "group" ||
			event.SubType != "normal" {
			return false
		}
		if !groupNums[event.GroupID] {
			return false
		}
		msg := new(coolq.Message)
		err := coolq.Unmarshal([]byte(event.Message), msg)
		if err != nil {
			logger.Field(_plugin.Name()).Error(err.Error())
			return false
		}
		sec := (*msg)[0]
		txt := sec.Data["text"]
		if sec.Type == "text" && strings.HasPrefix(txt, "# ") {
			return gameStarted[event.GroupID] && validPosInput[txt[2:]]
		}
		return false
	}
	filters["tic-tac-toe-game-end"] = func(event *coolq.CQEvent) bool {
		if event.PostType != "message" ||
			event.MessageType != "group" ||
			event.SubType != "normal" {
			return false
		}
		if !groupNums[event.GroupID] {
			return false
		}
		msg := new(coolq.Message)
		err := coolq.Unmarshal([]byte(event.Message), msg)
		if err != nil {
			logger.Field(_plugin.Name()).Error(err.Error())
			return false
		}
		sec := (*msg)[0]
		if sec.Type == "text" && sec.Data["text"] == "# 结束游戏" {
			return gameStarted[event.GroupID]
		}
		return false
	}
	return filters
}

// -1 未结束
// 0 平局
// 1 对手赢
// 2 晴乃赢
func checkWin(groupID int64) int {
	board := gameBoards[groupID]
	for i := 0; i < 3; i++ {
		if board[i][0] == board[i][1] && board[i][1] == board[i][2] && board[i][0] != 0 {
			if board[i][0] == 1 {
				return 1
			}
			return 2
		}
	}
	for i := 0; i < 3; i++ {
		if board[0][i] == board[1][i] && board[1][i] == board[2][i] && board[0][i] != 0 {
			if board[0][i] == 1 {
				return 1
			}
			return 2
		}
	}
	if board[0][0] == board[1][1] && board[1][1] == board[2][2] && board[0][0] != 0 {
		if board[0][0] == 1 {
			return 1
		}
		return 2
	}
	if board[0][2] == board[1][1] && board[2][0] == board[1][1] && board[0][2] != 0 {
		if board[0][2] == 1 {
			return 1
		}
		return 2
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == 0 {
				return -1
			}
			if i == 2 && j == 2 {
				return 0
			}
		}
	}
	return -1
}

func pick(groupID int64) (int, int) {
	board := gameBoards[groupID]
	retI := -1
	retJ := -1
	mx := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] != 0 {
				gameWeight[groupID][i][j] = 0
			} else {
				// 行和列
				// 最高权值
				if board[0][j]+board[1][j]+board[2][j] == 4 &&
					board[0][j]*board[1][j]*board[2][j] == 0 &&
					(board[0][j]-1)*(board[1][j]-1)*(board[2][j]-1) == -1 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10000
				}
				if board[i][0]+board[i][1]+board[i][2] == 4 &&
					board[i][0]*board[i][1]*board[i][2] == 0 &&
					(board[i][0]-1)*(board[i][1]-1)*(board[i][2]-1) == -1 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10000
				}
				// 次级权值
				if board[0][j]+board[1][j]+board[2][j] == 2 &&
					board[0][j]*board[1][j]*board[2][j] == 0 &&
					(board[0][j]-1)*(board[1][j]-1)*(board[2][j]-1) == 0 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 1000
				}
				if board[i][0]+board[i][1]+board[i][2] == 2 &&
					board[i][0]*board[i][1]*board[i][2] == 0 &&
					(board[i][0]-1)*(board[i][1]-1)*(board[i][2]-1) == 0 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 1000
				}
				// 三级权值（一排只有一个X）
				if board[0][j]+board[1][j]+board[2][j] == 1 &&
					board[0][j]*board[1][j]*board[2][j] == 0 &&
					(board[0][j]-1)*(board[1][j]-1)*(board[2][j]-1) == 0 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10
				}
				if board[i][0]+board[i][1]+board[i][2] == 1 &&
					board[i][0]*board[i][1]*board[i][2] == 0 &&
					(board[i][0]-1)*(board[i][1]-1)*(board[i][2]-1) == 0 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10
				}
				// 四级权值（一排只有一个O）
				if board[0][j]+board[1][j]+board[2][j] == 2 &&
					board[0][j]*board[1][j]*board[2][j] == 0 &&
					(board[0][j]-1)*(board[1][j]-1)*(board[2][j]-1) == 1 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 5
				}
				if board[i][0]+board[i][1]+board[i][2] == 2 &&
					board[i][0]*board[i][1]*board[i][2] == 0 &&
					(board[i][0]-1)*(board[i][1]-1)*(board[i][2]-1) == 1 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 5
				}
				// 五级权值（该行没有X或O）
				if board[0][j]+board[1][j]+board[2][j] == 0 &&
					board[0][j]*board[1][j]*board[2][j] == 0 &&
					(board[0][j]-1)*(board[1][j]-1)*(board[2][j]-1) == -1 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 2
				}
				if board[i][0]+board[i][1]+board[i][2] == 0 &&

					board[i][0]*board[i][1]*board[i][2] == 0 &&
					(board[i][0]-1)*(board[i][1]-1)*(board[i][2]-1) == -1 {
					gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 2
				}
				// 主对角线
				if i == 0 && j == 0 ||
					i == 2 && j == 2 ||
					i == 1 && j == 1 {
					// 最高权值
					if board[0][0]+board[1][1]+board[2][2] == 4 &&
						board[0][0]*board[1][1]*board[2][2] == 0 &&
						(board[0][0]-1)*(board[1][1]-1)*(board[2][2]-1) == -1 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10000
					}
					// 次级权值
					if board[0][0]+board[1][1]+board[2][2] == 2 &&
						board[0][0]*board[1][1]*board[2][2] == 0 &&
						(board[0][0]-1)*(board[1][1]-1)*(board[2][2]-1) == 0 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 1000
					}
					// 三级权值（一排只有一个X）
					if board[0][0]+board[1][1]+board[2][2] == 1 &&
						board[0][0]*board[1][1]*board[2][2] == 0 &&
						(board[0][0]-1)*(board[1][1]-1)*(board[2][2]-1) == 0 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10
					}
					// 四级权值（一排只有一个O）
					if board[0][0]+board[1][1]+board[2][2] == 2 &&
						board[0][0]*board[1][1]*board[2][2] == 0 &&
						(board[0][0]-1)*(board[1][1]-1)*(board[2][2]-1) == 1 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 5
					}
					// 五级权值（该行没有X或O）
					if board[0][0]+board[1][1]+board[2][2] == 0 &&
						board[0][0]*board[1][1]*board[2][2] == 0 &&
						(board[0][0]-1)*(board[1][1]-1)*(board[2][2]-1) == -1 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 2
					}
				}
				// 副对角线
				if i == 0 && j == 2 ||
					i == 2 && j == 0 ||
					i == 1 && j == 1 {
					// 最高权值
					if board[0][2]+board[1][1]+board[2][0] == 4 &&
						board[0][2]*board[1][1]*board[2][0] == 0 &&
						(board[0][2]-1)*(board[1][1]-1)*(board[2][0]-1) == -1 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10000
					}
					// 次级权值
					if board[0][2]+board[1][1]+board[2][0] == 2 &&
						board[0][2]*board[1][1]*board[2][0] == 0 &&
						(board[0][2]-1)*(board[1][1]-1)*(board[2][0]-1) == 0 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 1000
					}
					// 三级权值（一排只有一个X）
					if board[0][2]+board[1][1]+board[2][0] == 1 &&
						board[0][2]*board[1][1]*board[2][0] == 0 &&
						(board[0][2]-1)*(board[1][1]-1)*(board[2][0]-1) == 0 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 10
					}
					// 四级权值（一排只有一个O）
					if board[0][2]+board[1][1]+board[2][0] == 2 &&
						board[0][2]*board[1][1]*board[2][0] == 0 &&
						(board[0][2]-1)*(board[1][1]-1)*(board[2][0]-1) == 1 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 5
					}
					// 五级权值（该行没有X或O）
					if board[0][2]+board[1][1]+board[2][0] == 0 &&
						board[0][2]*board[1][1]*board[2][0] == 0 &&
						(board[0][2]-1)*(board[1][1]-1)*(board[2][0]-1) == -1 {
						gameWeight[groupID][i][j] = gameWeight[groupID][i][j] + 2
					}
				}
			}
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if gameWeight[groupID][i][j] > mx {
				mx = gameWeight[groupID][i][j]
				retI = i
				retJ = j
			}
		}
	}
	return retI, retJ
}

// Handlers 处理酷Q上报事件用
func (_plugin Game) Handlers() map[string]coolq.Handler {
	handlers := make(map[string]coolq.Handler)
	handlers["tic-tac-toe-game-start"] = func(event *coolq.CQEvent) {
		reply := coolq.NewMessage()
		groupID := event.GroupID
		if gameStarted[groupID] {
			reply = append(reply, coolq.NewTextSection("游戏已经开始！"))
			reply = append(reply, coolq.NewTextSection("请结束当前游戏"))
			reply = append(reply, coolq.NewTextSection("或输入\"# 结束游戏\"以结束游戏"))
			coolq.Client.SendGroupMsg(groupID, string(coolq.Marshal(reply)))
		} else {
			gameStarted[groupID] = true
			resetGameBoard(groupID)
			reply = append(reply, coolq.NewTextSection("游戏开始！"))
			reply = append(reply, coolq.NewTextSection("晴乃: O 对手: X"))
			reply = displayGameBoard(groupID, reply)
			reply = append(reply, coolq.NewTextSection("请以\"#\"开始，并输入坐标"))
			reply = append(reply, coolq.NewTextSection("例如 # A1,# B2,# C3"))
			coolq.Client.SendGroupMsg(groupID, string(coolq.Marshal(reply)))
		}
	}
	handlers["tic-tac-toe-gaming"] = func(event *coolq.CQEvent) {
		groupID := event.GroupID
		msg := new(coolq.Message)
		err := coolq.Unmarshal([]byte(event.Message), msg)
		if err != nil {
			logger.Field(_plugin.Name()).Error(err.Error())
		}
		sec := (*msg)[0]
		ipt := sec.Data["text"][2:]
		j := int(ipt[0] - 'A')
		i := int(ipt[1] - '1')
		if gameBoards[groupID][i][j] != 0 {
			coolq.Client.SendGroupMsg(groupID, "操作无效，请重试！")
			return
		}
		gameBoards[groupID][i][j] = 1
		reply := coolq.NewMessage()
		resTxt := ""
		res := checkWin(groupID)
		if res != -1 {
			gameStarted[event.GroupID] = false
		}
		switch res {
		case 0:
			resTxt = "平局！"
		case 1:
			resTxt = "你赢了！"
		case 2:
			resTxt = "是晴乃赢了！"
		}
		if len(resTxt) == 0 {
			rI, rJ := pick(groupID)
			if rI != -1 && rJ != -1 {
				gameBoards[groupID][rI][rJ] = 2
				res = checkWin(groupID)
				if res != -1 {
					gameStarted[event.GroupID] = false
				}
				switch res {
				case 0:
					resTxt = "平局！"
				case 1:
					resTxt = "你赢了！"
				case 2:
					resTxt = "是晴乃赢了！"
				}
			}
		}
		reply = append(reply, coolq.NewTextSection("井字棋游戏"))
		reply = append(reply, coolq.NewTextSection("晴乃: O 对手: X"))
		reply = displayGameBoard(groupID, reply)
		if len(resTxt) != 0 {
			reply = append(reply, coolq.NewTextSection(fmt.Sprintf("游戏结束: %s", resTxt)))
		}
		coolq.Client.SendGroupMsg(groupID, string(coolq.Marshal(reply)))
	}
	handlers["tic-tac-toe-game-end"] = func(event *coolq.CQEvent) {
		gameStarted[event.GroupID] = false
		coolq.Client.SendGroupMsg(event.GroupID, "游戏结束！")
	}
	return handlers
}

// Load 加载插件
func (_plugin Game) Load() error {
	return _plugin.loadConfig()
}

// Loaded 加载完成
func (_plugin Game) Loaded() {
	logger.Field(_plugin.Name()).Info("已成功加载")
}

// Instance 实体
var Instance = Game{}
