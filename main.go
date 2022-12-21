package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	libraryConfig "liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/gui"
	"liberty-town/node/network"
	"liberty-town/node/network/api_code/api_code_http"
	"liberty-town/node/network/api_code/api_code_websockets"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/api_implementation/api_http"
	"liberty-town/node/network/api_implementation/api_websockets"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/settings"
	"liberty-town/node/store"
	"liberty-town/node/validator"
	"liberty-town/node/validator/validation/validation_type"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"validator/challenge_hcaptcha"
	"validator/config"
	"validator/config/arguments"
	"validator/network/methods"
	"validator/validator_store"
)

func main() {
	fmt.Println("Starting Validator")
	var err error

	if err = arguments.InitArguments(os.Args[1:]); err != nil {
		return
	}
	if err = gui.InitGUI(); err != nil {
		return
	}
	if err = libraryConfig.InitConfig(); err != nil {
		return
	}

	if err = store.InitDB(); err != nil {
		panic(err)
	}

	//APIs
	api_http.ConfigureAPIRoutes = func(api *api_http.API) {
		api.GetMap = map[string]func(values url.Values) (interface{}, error){
			"":     api.GetMap[""],
			"ping": api.GetMap["ping"],
		}
		api.PostMap = map[string]func(values io.ReadCloser) (interface{}, error){
			"check":    api_code_http.HandlePOST[api_types.ValidatorCheckRequest, api_types.ValidatorCheckResult](methods.MethodCheck),
			"solution": api_code_http.HandlePOST[api_types.ValidatorSolutionRequest, api_types.ValidatorSolutionResult](methods.MethodSolution),
		}
		if config.CHALLENGE_TYPE == validation_type.VALIDATOR_CHALLENGE_CUSTOM {
			api.PostMap["challenge-custom"] = api_code_http.HandlePOST[methods.APIChallengeRequest, methods.APIChallengeCustomReply](methods.MethodChallengeCustom)
		}
	}

	api_websockets.ConfigureAPIRoutes = func(api *api_websockets.APIWebsockets) {
		api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) (interface{}, error){
			"":         api.GetMap[""],
			"ping":     api.GetMap["ping"],
			"check":    api_code_websockets.Handle[api_types.ValidatorCheckRequest, api_types.ValidatorCheckResult](methods.MethodCheck),
			"solution": api_code_websockets.Handle[api_types.ValidatorSolutionRequest, api_types.ValidatorSolutionResult](methods.MethodSolution),
		}
		if config.CHALLENGE_TYPE == validation_type.VALIDATOR_CHALLENGE_CUSTOM {
			api.GetMap["challenge-custom"] = api_code_websockets.Handle[methods.APIChallengeRequest, methods.APIChallengeCustomReply](methods.MethodChallengeCustom)
		}
	}

	//提供静态文件
	network_config.STATIC_FILES["/static/"] = "../../../static"

	if err = network.NewNetwork(); err != nil {
		return
	}

	if err = settings.Load(); err != nil {
		panic(err)
	}

	if err = validator_store.InitializeStore(); err != nil {
		panic(err)
	}

	config.CHALLENGE_0 = cryptography.RandomHash()

	if err = challenge_hcaptcha.Init(); err != nil {
		panic(err)
	}

	myValidator := &validator.Validator{
		validator.VALIDATOR_VERSION,
		settings.Settings.Load().Contact.Contact,
		&ownership.Ownership{},
	}
	if err = myValidator.Ownership.Sign(settings.Settings.Load().Account.PrivateKey, myValidator.GetMessageToSign); err != nil {
		panic(err)
	}

	b, err := json.Marshal(myValidator)
	if err != nil {
		panic(err)
	}

	//调试信息
	gui.GUI.Info("VALIDATOR JSON", string(b))
	gui.GUI.Info("VALIDATOR JSON OWNERSHIP", base64.StdEncoding.EncodeToString(myValidator.Ownership.SerializeToBytes()))

	exitCn := make(chan os.Signal, 10)
	signal.Notify(exitCn,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGSTOP,
	)

	<-exitCn

	signal.Stop(exitCn)
}
