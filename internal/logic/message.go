package logic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"tgwp/global"
	"tgwp/internal/repo"
	"tgwp/internal/types"
	"tgwp/log/zlog"
	"tgwp/util/snowflake"
)

type MessageLogic struct {
}

func NewMessageLogic() *MessageLogic {
	return &MessageLogic{}
}

func (l *MessageLogic) SetMessage(c *gin.Context, req types.SetMessageReq) (resp types.SetMessageResp, err error) {
	// 生成雪花ID生成器
	node, err := snowflake.NewNode(global.DEFAULT_NODE_ID)
	if err != nil {
		zlog.Errorf("create snowflake node error:%v", err)
		return
	}
	// 生成雪花ID
	id := node.Generate().Int64()

	message, err := repo.NewMessageRepo(global.DB).CreateMessage(id, req.Content, req.Type)
	if err != nil {
		zlog.Errorf("create message error:%v", err)
		return
	} else {
		zlog.Infof("create message success, id:%d , content:\"%s\", type:%d", message.ID, message.Content, message.Type)
	}

	resp.MessageID = id

	return
}

func (l *MessageLogic) JoinMessage(c *gin.Context, req types.JoinMessageReq, UserID string) (resp types.JoinMessageResp, err error) {
	// 生成雪花ID生成器
	node, err := snowflake.NewNode(global.DEFAULT_NODE_ID)
	if err != nil {
		zlog.Errorf("create snowflake node error:%v", err)
		return
	}

	// 判断信息是否存在
	//is_exist := repo.NewMessageRepo(global.DB).CheckMessageExist(req.MessageID)
	//if is_exist == false {
	//	zlog.Warnf("message not exist, message_id:%d", req.MessageID)
	//	err = fmt.Errorf("message not exist")
	//	return
	//}

	// 生成雪花ID
	id := node.Generate().Int64()

	// 更新数据库
	user_message, err := repo.NewMessageRepo(global.DB).CreateUserMessage(id, req.MessageID, UserID)
	if err != nil {
		zlog.Errorf("create message error:%v", err)
		return
	} else {
		zlog.Infof("create message success, user_id:%d, message_id:%d", user_message.UserID, user_message.MessageID)
	}

	return
}

func (l *MessageLogic) GetMessage(c *gin.Context, UserID string, pageStr string, timestampStr string) (resp types.GetMessageResp, err error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return
	}
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return
	}
	fmt.Println(timestamp)

	is_update := repo.NewMessageRepo(global.DB).CheckUpdate(UserID, timestamp)
	if is_update {
		resp, err = repo.NewMessageRepo(global.DB).GetMessage(UserID, page, 5)
		resp.IsUpdated = true
	} else {
		resp.IsUpdated = false
	}

	if err != nil {
		zlog.Errorf("get message error:%v", err)
		return
	} else {
		zlog.Infof("get message success")
	}

	return
}

func (l *MessageLogic) MarkReadMessage(c *gin.Context, req types.MarkReadMessageReq) (resp types.JoinMessageResp, err error) {
	// 判断信息是否存在
	//is_exist := repo.NewMessageRepo(global.DB).CheckUserMessageExist(req.UserMessageID)
	//if is_exist == false {
	//	zlog.Warnf("message not exist, message_id:%d", req.UserMessageID)
	//	err = fmt.Errorf("message not exist")
	//	return
	//}

	// 更新数据库
	err = repo.NewMessageRepo(global.DB).MarkReadMessage(req.UserMessageID)

	if err != nil {
		zlog.Errorf("mark read error:%v", err)
		return
	} else {
		zlog.Infof("mark read success, user_message_id:%d", req.UserMessageID)
	}

	return
}

func (l *MessageLogic) SendMessage(req types.SendMessageReq, UserID int64) (resp types.SendMessageResp, err error) {
	// 使用 SetMessage 方法
	respSet, err := l.SetMessage(types.SetMessageReq{
		Content: req.Content,
		Type:    req.Type,
	})
	if err != nil {
		return
	}
	// 使用 JoinMessage 方法
	respJoin, err := l.JoinMessage(types.JoinMessageReq{
		MessageID: respSet.MessageID,
	}, UserID)
	if err != nil {
		return
	}

	resp.MessageID = respSet.MessageID
	resp.UserMessageID = respJoin.UserMessageID
	return
}
