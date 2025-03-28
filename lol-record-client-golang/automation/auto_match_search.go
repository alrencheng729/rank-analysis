package automation

import (
	"context"
	"lol-record-analysis/common/config"
	"lol-record-analysis/lcu/client/api"
	"lol-record-analysis/lcu/client/constants"
	"lol-record-analysis/util/init_log"
	"time"
)

// 添加一个标志位来控制是否自动匹配
var autoMatchEnabled = true

// 添加一个变量来存储上一次的匹配状态
var lastSearchState string

func startMatchAutomation(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			// 获取当前匹配状态
			curState, err := api.GetPhase()
			if err != nil {
				init_log.AppLog.Error(err.Error())
				continue
			}
			// 检测匹配状态变化
			if lastSearchState == curState {
				continue
			}
			// 检查是否开启自动匹配
			if !autoMatchEnabled {
				continue
			}

			// 检查配置中的自动匹配开关
			if !config.Viper().GetBool("settings.auto.startMatchSwitch") {
				continue
			}
			// 取消后关闭
			if lastSearchState == constants.Matchmaking && curState == constants.Lobby {
				autoMatchEnabled = false
			}
			// 开始后启动
			if lastSearchState == constants.Lobby && curState == constants.Matchmaking {
				autoMatchEnabled = true
			}
			lastSearchState = curState

			// 检查当前游戏阶段
			curPhase, err := api.GetPhase()
			if err != nil {
				init_log.AppLog.Error(err.Error())
				continue
			}
			if curPhase != "Lobby" {
				continue
			}

			// 获取房间信息
			lobby, err := api.GetLobby()
			if err != nil {
				init_log.AppLog.Error(err.Error())
				continue
			}

			// 检查是否是自定义游戏
			if lobby.GameConfig.IsCustom {
				continue
			}

			// 检查是否是房主
			if len(lobby.Members) > 0 && !lobby.Members[0].IsLeader {
				continue
			}

			// 开始匹配
			api.PostMatchSearch()
		}
	}
}
