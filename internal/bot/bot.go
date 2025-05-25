package bot

import (
	"db"
	"fmt"
	"github"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	owner = "iamhalje"
	repo  = "budgetbot"
)

type Bot struct {
	BotAPI *tgbotapi.BotAPI
	DB     *db.DB
}

func NewBot(botAPI *tgbotapi.BotAPI, database *db.DB) *Bot {
	return &Bot{
		BotAPI: botAPI,
		DB:     database,
	}
}

func (b *Bot) HandleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.BotAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("Получено сообщение от пользователя %d: %q", update.Message.From.ID, update.Message.Text)
		b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	userID := msg.From.ID
	text := msg.Text

	log.Printf("Обработка сообщения от %d: %q", userID, text)

	user, err := db.GetUserByTelegramID(b.DB.DB, int64(userID))
	if err != nil {
		log.Printf("Ошибка получения пользователя %d: %v", userID, err)
		b.reply(msg.Chat.ID, "Ошибка при получении данных пользователя")
		return
	}
	if user == nil {
		log.Printf("Пользователь %d не зарегистрирован", userID)
		b.reply(msg.Chat.ID, "Пользователь не зарегистрирован, используйте /login для регистрации")
		return
	}

	monthReset, err := db.ResetIfNewMonth(b.DB.DB, user)
	if err != nil {
		log.Printf("Ошибка сброса данных по новому месяцу для пользователя %d: %v", userID, err)
	} else if monthReset {
		log.Printf("Новый месяц, сброс трат для пользователя %d", userID)
		b.reply(msg.Chat.ID, "Наступил новый месяц, траты обнулены!")
	}

	switch {
	case strings.HasPrefix(text, "/login"):
		log.Printf("Команда /login от пользователя %d", userID)
		b.cmdLogin(msg, userID, text)
	case strings.HasPrefix(text, "/setbudget"):
		log.Printf("Команда /setbudget от пользователя %d", userID)
		b.cmdSetBudget(msg, userID, text)
	case strings.HasPrefix(text, "/spend"):
		log.Printf("Команда /spend от пользователя %d", userID)
		b.cmdSpend(msg, userID, text)
	case text == "/help":
		log.Printf("Команда /help от пользователя %d", userID)
		b.cmdHelp(msg)
	case strings.HasPrefix(text, "/resetspent"):
		log.Printf("Команда /resetspent от пользователя %d", userID)
		b.cmdResetSpent(msg, userID)
	default:
		log.Printf("Неизвестная команда от пользователя %d: %q", userID, text)
		b.reply(msg.Chat.ID, "Команда не определена, используйте /help для вывода списка команд")
	}
}

func (b *Bot) cmdLogin(msg *tgbotapi.Message, userID int64, text string) {
	args := strings.Fields(text)
	if len(args) != 2 {
		log.Printf("Неверный формат /login от %d: %q", userID, text)
		b.reply(msg.Chat.ID, "Используйте: /login github_username")
		return
	}
	githubLogin := args[1]

	exists, err := db.ExistsGithubLogin(b.DB.DB, githubLogin)
	if err != nil {
		log.Printf("Ошибка базы данных при проверке githubLogin %q: %v", githubLogin, err)
		b.reply(msg.Chat.ID, "Ошибка базы данных")
		return
	}
	if exists {
		log.Printf("GitHub пользователь %q уже зарегистрирован", githubLogin)
		b.reply(msg.Chat.ID, "Этот GitHub пользователь уже зарегистрирован")
		return
	}

	stargazers, err := github.GetStargazers(owner, repo)
	if err != nil {
		log.Printf("Ошибка получения stargazers GitHub: %v", err)
		b.reply(msg.Chat.ID, "Ошибка проверки GitHub")
		return
	}
	if !github.IsUserStargazer(githubLogin, stargazers) {
		log.Printf("Доступ запрещён для пользователя %q - нет звезды", githubLogin)
		b.reply(msg.Chat.ID, "Доступ запрещен, вы не поставили звезду на репозиторий.\nПожалуйста, поставьте ⭐ на репозиторий: https://github.com/iamhalje/budgetbot")
		return
	}

	err = db.UpdateUser(b.DB.DB, userID, githubLogin)
	if err != nil {
		log.Printf("Ошибка регистрации пользователя %d с github %q: %v", userID, githubLogin, err)
		b.reply(msg.Chat.ID, "Ошибка базы данных при регистрации")
		return
	}

	log.Printf("Пользователь %d успешно зарегистрирован с github %q", userID, githubLogin)
	b.reply(msg.Chat.ID, "Регистрация прошла успешно, теперь вы можете установить месячный бюджет командой /setbudget")
}

func (b *Bot) cmdSetBudget(msg *tgbotapi.Message, userID int64, text string) {
	user, err := db.GetUserByTelegramID(b.DB.DB, userID)
	if err != nil || user == nil {
		b.reply(msg.Chat.ID, "Сначала выполните регистрацию через /login github_username")
		return
	}

	args := strings.Fields(text)
	if len(args) != 2 {
		b.reply(msg.Chat.ID, "Используйте: /setbudget сумма")
		return
	}

	budget, err := strconv.ParseFloat(args[1], 64)
	if err != nil || budget <= 0 {
		b.reply(msg.Chat.ID, "Введите корректное число бюджета")
		return
	}

	nowMonth := time.Now().Format("2006-01")

	err = db.SetBudget(b.DB.DB, userID, budget, nowMonth)
	if err != nil {
		b.reply(msg.Chat.ID, "Ошибка записи бюджета")
		return
	}

	b.reply(msg.Chat.ID, fmt.Sprintf("Бюджет установлен: %.2f на месяц %s", budget, nowMonth))
}

func (b *Bot) cmdSpend(msg *tgbotapi.Message, userID int64, text string) {
	user, err := db.GetUserByTelegramID(b.DB.DB, userID)
	if err != nil || user == nil {
		b.reply(msg.Chat.ID, "Сначала выполните регистрацию через /login github_username")
		return
	}

	args := strings.Fields(text)
	if len(args) != 2 {
		b.reply(msg.Chat.ID, "Используйте: /spend сумма")
		return
	}

	spent, err := strconv.ParseFloat(args[1], 64)
	if err != nil || spent <= 0 {
		b.reply(msg.Chat.ID, "Введите корректную сумму расхода")
		return
	}

	newSpent := user.Spent + spent

	if newSpent > user.MonthlyBudget {
		b.reply(msg.Chat.ID, fmt.Sprintf("Внимание, вы превысили бюджет на %.2f", newSpent-user.MonthlyBudget))
	}

	err = db.UpdateSpent(b.DB.DB, userID, newSpent)
	if err != nil {
		b.reply(msg.Chat.ID, "Ошибка обновления расходов")
		return
	}

	b.reply(msg.Chat.ID, fmt.Sprintf("Расходны обновлены, всего потрачено %.2f из %.2f", newSpent, user.MonthlyBudget))
}

func (b *Bot) cmdResetSpent(msg *tgbotapi.Message, userID int64) {
	user, err := db.GetUserByTelegramID(b.DB.DB, userID)
	if err != nil || user == nil {
		b.reply(msg.Chat.ID, "Сначала выполните регистрацию через /login github_username")
		return
	}

	nowMonth := time.Now().Format("2006-01")
	if user.BudgetMonth != nowMonth {
		b.reply(msg.Chat.ID, "У вас нет активного бюджета на этот месяц, установите его через /setbudget")
		return
	}

	err = db.UpdateSpent(b.DB.DB, userID, 0)
	if err != nil {
		b.reply(msg.Chat.ID, "Ошибка при сбросе расходов")
		return
	}

	b.reply(msg.Chat.ID, "Расходы успешно сброшены на 0 для текущего месяца")
}

func (b *Bot) cmdHelp(msg *tgbotapi.Message) {
	text := `/login github_username - зарегистрироваться через GitHub
/setbudget сумма - установить месячный бюджет
/spend сумма - добавить сумму расхода
/resetspent — сбросить расходы текущего месяца
/help - показать эту справку

Если возникли проблемы или есть предложения, пожалуйста, создайте issue в репозитории:
https://github.com/iamhalje/MonthlyBudgetBot/issues`
	b.reply(msg.Chat.ID, text)
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.BotAPI.Send(msg)
	if err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}
