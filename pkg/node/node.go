package node

import (
	"net/http"
	"strings"

	"errors"
	"fmt"

	common "../common"
	e "../e"
	model "../model"
	request "../request"
	response "../response"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	NodeType string `json:"node_type" binding:"required"`
}

var node *Node

func getNode() *Node {
	if node != nil {
		return node
	}
	node = new(Node)
	return node
}

// .. // User

func (node Node) CreatePlayer(c *gin.Context) {

	var createPlayerRequest request.CreatePlayerRequest
	err := c.BindJSON(&createPlayerRequest)
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		createPlayerRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
					Status: 0, Msg: err.Error()})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
						Status: 0, Msg: err.Error()})
					return
				}
			}

			if createPlayerRequest.AgentID != tokens[0].AgentID.Hex() {
				err := errors.New("Auth Token doesn't match with the agent")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
						Status: 0, Msg: err.Error()})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
					Status: 0, Msg: err.Error()})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
				Status: 0, Msg: err.Error()})
			return
		}
	}

	agents, err := model.GetAgentByIDFromDB(createPlayerRequest.AgentID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(agents) == 0 {
		err := errors.New("Bad Agent ID")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	wallet, err := model.CreateWalletnInDB()

	player := model.Player{AgentID: bson.ObjectIdHex(createPlayerRequest.AgentID),
		UserName:  createPlayerRequest.UserName,
		FirstName: createPlayerRequest.FirstName,
		LastName:  createPlayerRequest.LastName,
		Email:     createPlayerRequest.Email,
		Gender:    createPlayerRequest.Gender,
		CellPhone: createPlayerRequest.CellPhone,
		BirthDate: createPlayerRequest.BirthDate,
		WalletID:  wallet.ID,
		//Currency:  createPlayerRequest.Currency,
		//CountryID: createPlayerRequest.CountryID//
	}

	player, err = model.CreatePlayerInDb(player)

	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.CreatePlayerResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.CreatePlayerResponse{
		AgentName: agents[0].AgentName,
		UserName:  player.UserName,
		FirstName: player.FirstName,
		LastName:  player.LastName,
		CellPhone: player.CellPhone,
		Email:     player.Email,
		Gender:    player.Gender,
		BirthDate: player.BirthDate,
		//Currency:  player.Currency,
		//CountryID: player.CountryID,
		Status:   1,
		ID:       fmt.Sprintf("%x", string(player.ID)),
		WalletID: fmt.Sprintf("%x", string(wallet.ID)),
	})
	return
}

//... // Agent

func (node Node) GetAgentToken(c *gin.Context) {
	username, password := c.Query("username"), c.Query("password")
	getTokenRequest := request.GetTokenRequest{UserName: username, Password: password}
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request -",
		getTokenRequest)

	agents, err := model.GetAgentFromDB(username, password)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.GetTokenResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(agents) == 0 {
		err := errors.New("Bad Username/Password/Failed to Authenticate")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.GetTokenResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	token, err := model.CreateTokenInDB(agents[0].ID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.GetTokenResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.GetTokenResponse{
		Status:   1,
		Token:    "Bearer " + token.TokenString,
		ExpireAt: common.GetTimeInAllInt(model.GetExpireDate(token.CreateAt, model.ExpireSecond))})
}

func (node Node) GetGameString(c *gin.Context) {
	locale := c.Query("locale")
	getGameStringRequest := request.GetGameStringRequest{Locale: locale}
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request -",
		getGameStringRequest)

	games, err := model.GetGameFromDB(locale)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, games)
}

func (node Node) LaunchGame(c *gin.Context) {
	var launchGameRequest request.LaunchGameRequest
	err := c.BindJSON(&launchGameRequest)
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		launchGameRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
			return
		}
	}

	games, err := model.GetGameByIDFromDB(launchGameRequest.GameID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
		return
	}

	if len(games) == 0 {
		err := errors.New("Bad Game id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
		return
	}

	players, err := model.GetPlayerByIDFromDB(launchGameRequest.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
		return
	}

	if len(players) == 0 {
		err := errors.New("Bad Player id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
		return
	}

	play, err := model.CreatePlayInDB(model.Play{PlayerID: players[0].ID, GameID: games[0].ID, ReturnURL: launchGameRequest.ReturnURL})
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.LaunchGameResponse{Status: 0, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, response.LaunchGameResponse{Status: 1, Token: play.TokenString,
		PlayURL: common.GetConfiger().Configs.GameNodeAddress + "?game_id=" + fmt.Sprintf("%x", string(play.GameID)) +
			"&player_id=" + fmt.Sprintf("%x", string(play.PlayerID)) +
			"&token=" + play.TokenString +
			"&return_url=" + launchGameRequest.ReturnURL})
}

func (node Node) Deposit(c *gin.Context) {
	var depositRequest request.DepositRequest
	err := c.BindJSON(&depositRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.DepositResponse{
					Status: 0,
					Msg:    err.Error(),
				})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusInternalServerError, response.DepositResponse{
						Status: 0,
						Msg:    err.Error(),
					})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.DepositResponse{
					Status: 0,
					Msg:    err.Error(),
				})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusInternalServerError, response.DepositResponse{
				Status: 0,
				Msg:    err.Error(),
			})
			return
		}
	}

	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		depositRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	players, err := model.GetPlayerByIDFromDB(depositRequest.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	if len(players) == 0 {
		err := errors.New("Bad Player id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	wallets, err := model.GetWalletByIDFromDB(players[0].WalletID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	if len(wallets) == 0 {
		err := errors.New("No Wallet Found")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	if depositRequest.Money < 0 {
		err := errors.New("Deposit Money cannot be negative")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	remainMoney, err := model.UpdateWalletMoneyInDB(wallets[0].ID, depositRequest.Money)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	trans, err := model.CreateTransactionInDB(model.Transaction{MoneyExchange: depositRequest.Money,
		MoneyRemain: remainMoney, PlayerID: players[0].ID, TransIDPlatform: depositRequest.TransIDPlatform})

	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.DepositResponse{
		Status:        1,
		PlayerID:      depositRequest.PlayerID,
		TransactionID: fmt.Sprintf("%x", string(trans.ID))})

	return
}

func (node Node) GetBalance(c *gin.Context) {
	var balanceRequest request.BalanceRequest
	err := c.BindJSON(&balanceRequest)
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		&balanceRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
					Status: 0, Msg: err.Error()})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
						Status: 0, Msg: err.Error()})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
					Status: 0, Msg: err.Error()})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
				Status: 0, Msg: err.Error()})
			return
		}
	}

	players, err := model.GetPlayerByIDFromDB(balanceRequest.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(players) == 0 {
		err := errors.New("Bad Player id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	wallets, err := model.GetWalletByIDFromDB(players[0].WalletID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(wallets) == 0 {
		err := errors.New("No Wallet Found")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetBalanceResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.GetBalanceResponse{
		Status: 1,
		Amount: wallets[0].Money})
	return
}

func (node Node) WithDraw(c *gin.Context) {
	var withDrawRequest request.WithDrawRequest
	err := c.BindJSON(&withDrawRequest)
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		withDrawRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.DepositResponse{
					Status: 0,
					Msg:    err.Error(),
				})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusInternalServerError, response.DepositResponse{
						Status: 0,
						Msg:    err.Error(),
					})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusInternalServerError, response.DepositResponse{
					Status: 0,
					Msg:    err.Error(),
				})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusInternalServerError, response.DepositResponse{
				Status: 0,
				Msg:    err.Error(),
			})
			return
		}
	}

	players, err := model.GetPlayerByIDFromDB(withDrawRequest.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	if len(players) == 0 {
		err := errors.New("Bad Player id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	wallets, err := model.GetWalletByIDFromDB(players[0].WalletID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	if len(wallets) == 0 {
		err := errors.New("No Wallet Found")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	if withDrawRequest.Money < 0 {
		err := errors.New("WithDraw Money cannot be negative")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	remainMoney, err := model.UpdateWalletMoneyInDB(wallets[0].ID, -withDrawRequest.Money)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	trans, err := model.CreateTransactionInDB(model.Transaction{MoneyExchange: -withDrawRequest.Money,
		MoneyRemain: remainMoney, PlayerID: players[0].ID, TransIDPlatform: withDrawRequest.TransIDPlatform})
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusInternalServerError, response.DepositResponse{
			Status: 0,
			Msg:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.DepositResponse{Status: 1, PlayerID: withDrawRequest.PlayerID, TransactionID: fmt.Sprintf("%x", string(trans.ID))})
}

func (node Node) GetTransactionByPID(c *gin.Context) {
	var getTransByPIDRequest request.GetTransByPIDRequest
	err := c.BindJSON(&getTransByPIDRequest)
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		&getTransByPIDRequest)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
					Status: 0, Msg: err.Error()})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
						Status: 0, Msg: err.Error()})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
					Status: 0, Msg: err.Error()})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
				Status: 0, Msg: err.Error()})
			return
		}
	}

	players, err := model.GetPlayerByIDFromDB(getTransByPIDRequest.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(players) == 0 {
		err := errors.New("Bad Player id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	trans, err := model.GetTransactionByIDFromDB(getTransByPIDRequest.PlayerID, getTransByPIDRequest.TransIDPlatform)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(trans) == 0 {
		err := errors.New("No Transaction Found")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetTransByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.GetTransByPIDResponse{
		Status:        1,
		MoneyExchange: trans[0].MoneyExchange,
		MoneyRemain:   trans[0].MoneyRemain,
		TransactionID: fmt.Sprintf("%x", string(trans[0].ID))})

	return
}

func (node Node) GetGameRecordByPID(c *gin.Context) {
	var getGameRecordByPID request.GetGameRecordByPIDRequest
	err := c.BindJSON(&getGameRecordByPID)
	common.GetLogger().Log(e.TRACK, "Request Host - ", common.GetCurrentIP(*c.Request), "|", "Request JSON -",
		&getGameRecordByPID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	value := c.GetHeader("Authorization")
	if value != "" {
		n := strings.Index(value, "Bearer ")
		if n > -1 {
			tokenStringStartIndex := n + len("Bearer ")
			tokens, err := model.GetTokenFromDB(value[tokenStringStartIndex:len(value)])
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
					Status: 0, Msg: err.Error()})
				return
			}

			if len(tokens) == 0 {
				err := errors.New("No Tokens Found")
				if err != nil {
					common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
						err.Error())
					c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
						Status: 0, Msg: err.Error()})
					return
				}
			}
		} else {
			err := errors.New("Invalid Auth Token")
			if err != nil {
				common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
					err.Error())
				c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
					Status: 0, Msg: err.Error()})
				return
			}
		}
	} else {
		err := errors.New("Invalid Auth Token")
		if err != nil {
			common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
				err.Error())
			c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
				Status: 0, Msg: err.Error()})
			return
		}
	}

	players, err := model.GetPlayerByIDFromDB(getGameRecordByPID.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	if len(players) == 0 {
		err := errors.New("Bad Player id")
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	gameRecords, err := model.GetGameRecord(getGameRecordByPID.PlayerID)
	if err != nil {
		common.GetLogger().Log(e.ERROR, "Request Host -", common.GetCurrentIP(*c.Request), "|", "Error -",
			err.Error())
		c.JSON(http.StatusBadRequest, response.GetGameRecordByPIDResponse{
			Status: 0, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.GetGameRecordByPIDResponse{
		Status:      1,
		GameRecords: gameRecords,
	})

}
