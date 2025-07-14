package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"

	_ "github.com/go-sql-driver/mysql"

	nkAPIClient "game-server-paper-rock-scissors/module/nakama/nakama_custom_api_clients"

	nkRPC "game-server-paper-rock-scissors/module/nakama/nakama_custom_rpc"

	//customHttp "game-server-paper-rock-scissors/module/CustomHttp"
	customAPI "game-server-paper-rock-scissors/module/nakama/nakama_custom_api_server"
)

type MatchState struct {
	presences map[string]runtime.Presence
}

type Match struct{}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {

	//go customHttp.StartHosting()

	customAPI.RegisterCustomAPI(initializer)

	//Listen when a user logs into their account
	initializer.RegisterAfterAuthenticateEmail(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateEmailRequest) error {

		userToken := out.Token

		logger.Info("User Token: %v\n", userToken)

		userID := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)

		account, err := nk.AccountGetId(ctx, userID)

		if err != nil {
			logger.Info("Cannot get user ID: %v\n", err)
		}

		username := account.GetUser().Username

		logger.Info("user name: %v | user id: %v", username, userID)

		nkAPIClient.CallRegisterUserAPI(userToken, username, account.GetEmail())

		return nil
	})

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	logger.Info("Initializing Modules......")

	RegisterMatchErr := initializer.RegisterMatch("standard_match", newMatch)

	if RegisterMatchErr != nil {
		logger.Error("[RegisterMatch] error: ", RegisterMatchErr.Error())
		return RegisterMatchErr
	}

	rpcErr := initializer.RegisterRpc("RPCgetInventoryData", nkRPC.RPCgetInventoryData)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	rpcErr = initializer.RegisterRpc("RPCgetUserCredit", nkRPC.RPCgetUserCredit)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	rpcErr = initializer.RegisterRpc("RPCSpinForReward", nkRPC.RPCSpinForReward)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	rpcErr = initializer.RegisterRpc("RPCclearInventory", nkRPC.RPCclearInventory)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	rpcErr = initializer.RegisterRpc("RPCSaveUserQuest", nkRPC.RPCSaveUserQuest)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	rpcErr = initializer.RegisterRpc("RPCGetUserQuest", nkRPC.RPCGetUserQuest)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	rpcErr = initializer.RegisterRpc("RPCUpdateUserQuest", nkRPC.RPCUpdateUserQuest)

	if rpcErr != nil {
		return &runtime.Error{Message: fmt.Sprintf("Failed to register rpc: %v", rpcErr)}
	}

	return nil
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		presences: make(map[string]runtime.Presence),
	}

	tickRate := 1
	label := ""

	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	acceptUser := true

	return state, acceptUser, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		mState.presences[p.GetUserId()] = p
	}

	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	for _, p := range presences {
		delete(mState.presences, p.GetUserId())
	}

	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	for _, presence := range mState.presences {
		logger.Info("Presence %v named %v", presence.GetUserId(), presence.GetUsername())
	}

	for _, message := range messages {
		logger.Info("Received %v from %v", string(message.GetData()), message.GetUserId())
	}

	return mState
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {

	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, "signal received: " + data
}
