package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"xboard/internal/config"
	"xboard/internal/model"
)

// TelegramService Telegram Bot 服务
type TelegramService struct {
	botToken   string
	chatID     string
	httpClient *http.Client
}

func NewTelegramService(cfg config.TelegramConfig) *TelegramService {
	return &TelegramService{
		botToken:   cfg.BotToken,
		chatID:     cfg.ChatID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetBotToken 获取 Bot Token
func (s *TelegramService) GetBotToken() string {
	return s.botToken
}

// TelegramUpdate Telegram 更新
type TelegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

// TelegramMessage Telegram 消息
type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      *TelegramChat `json:"chat"`
	Text      string        `json:"text"`
	Date      int64         `json:"date"`
}

// TelegramUser Telegram 用户
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// TelegramChat Telegram 聊天
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// SendMessage 发送消息
func (s *TelegramService) SendMessage(chatID int64, text string, parseMode string) error {
	if s.botToken == "" {
		return fmt.Errorf("telegram bot not configured")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	params := url.Values{}
	params.Set("chat_id", fmt.Sprintf("%d", chatID))
	params.Set("text", text)
	if parseMode != "" {
		params.Set("parse_mode", parseMode)
	}

	resp, err := s.httpClient.PostForm(apiURL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram api error: %s", string(body))
	}

	return nil
}

// SendMarkdown 发送 Markdown 消息
func (s *TelegramService) SendMarkdown(chatID int64, text string) error {
	return s.SendMessage(chatID, text, "Markdown")
}

// SendHTML 发送 HTML 消息
func (s *TelegramService) SendHTML(chatID int64, text string) error {
	return s.SendMessage(chatID, text, "HTML")
}

// HandleUpdate 处理 Telegram 更新
func (s *TelegramService) HandleUpdate(update *TelegramUpdate) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	text := strings.TrimSpace(msg.Text)

	// 命令处理
	if strings.HasPrefix(text, "/") {
		return s.handleCommand(msg)
	}

	return nil
}

// handleCommand 处理命令
func (s *TelegramService) handleCommand(msg *TelegramMessage) error {
	parts := strings.Fields(msg.Text)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/start":
		return s.cmdStart(msg)
	case "/help":
		return s.cmdHelp(msg)
	default:
		return s.SendMessage(msg.Chat.ID, "未知命令，输入 /help 查看帮助", "")
	}
}

// cmdStart 开始命令
func (s *TelegramService) cmdStart(msg *TelegramMessage) error {
	text := `🎉 *欢迎使用 XBoard Bot*

使用以下命令管理您的账户：

/bind <邮箱> - 绑定账户
/unbind - 解绑账户
/info - 查看账户信息
/traffic - 查看流量使用
/subscribe - 获取订阅链接
/help - 帮助信息`

	return s.SendMarkdown(msg.Chat.ID, text)
}

// cmdHelp 帮助命令
func (s *TelegramService) cmdHelp(msg *TelegramMessage) error {
	text := `📖 *帮助信息*

*账户管理*
/bind <邮箱> - 绑定账户
/unbind - 解绑账户

*信息查询*
/info - 查看账户信息
/traffic - 查看流量使用
/subscribe - 获取订阅链接

*其他*
/help - 显示此帮助`

	return s.SendMarkdown(msg.Chat.ID, text)
}

// getUserStatus 获取用户状态
func (s *TelegramService) getUserStatus(user *model.User) string {
	if user.Banned {
		return "🚫 已封禁"
	}
	if !user.IsActive() {
		return "⏸️ 已过期"
	}
	return "✅ 正常"
}

// FormatBytes 格式化字节
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// NotifyExpire 通知用户到期
func (s *TelegramService) NotifyExpire(user *model.User, daysLeft int) error {
	if user.TelegramID == nil || *user.TelegramID == 0 {
		return nil
	}

	text := fmt.Sprintf(`⏰ *订阅到期提醒*

您的订阅将在 *%d 天*后到期，请及时续费以避免服务中断。`, daysLeft)

	return s.SendMarkdown(*user.TelegramID, text)
}

// NotifyTrafficWarning 通知流量预警
func (s *TelegramService) NotifyTrafficWarning(user *model.User, usedPercent int) error {
	if user.TelegramID == nil || *user.TelegramID == 0 {
		return nil
	}

	text := fmt.Sprintf(`📊 *流量使用提醒*

您的流量已使用 *%d%%*，请合理使用或考虑升级套餐。`, usedPercent)

	return s.SendMarkdown(*user.TelegramID, text)
}

// NotifyNewTicket 通知管理员新工单
func (s *TelegramService) NotifyNewTicket(subject, userEmail string) error {
	if s.chatID == "" {
		return nil
	}

	text := fmt.Sprintf(`🎫 *新工单*

用户：%s
主题：%s`, userEmail, subject)

	// 解析 chatID
	var chatID int64
	fmt.Sscanf(s.chatID, "%d", &chatID)
	if chatID == 0 {
		return nil
	}

	return s.SendMarkdown(chatID, text)
}

// SetWebhook 设置 Webhook
func (s *TelegramService) SetWebhook(webhookURL string) error {
	if s.botToken == "" {
		return fmt.Errorf("telegram bot not configured")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", s.botToken)

	data := map[string]string{"url": webhookURL}
	body, _ := json.Marshal(data)

	resp, err := s.httpClient.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set webhook failed: %s", string(respBody))
	}

	return nil
}
