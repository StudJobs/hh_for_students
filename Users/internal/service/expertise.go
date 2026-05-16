package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/google/uuid"

	"github.com/studjobs/hh_for_students/users/internal/repository"
)

// passThresholdPct — порог прохождения теста. >= = passed.
const passThresholdPct = 70

// staticQuestion — структура для вопроса в каталоге.
type staticQuestion struct {
	Text          string
	Options       []string
	CorrectOption int
}

// expertiseCatalog — статичный каталог тестов по навыкам. Простой набор
// тестов выбран сознательно: для MVP важна сама механика верификации эксперта,
// а не глубина проверки знаний.
var expertiseCatalog = map[string][]staticQuestion{
	"go": {
		{
			Text:          "Какая команда инициализирует Go-модуль?",
			Options:       []string{"go init", "go mod init", "go module new", "go new"},
			CorrectOption: 1,
		},
		{
			Text:          "Что делает горутина?",
			Options:       []string{"Управляет памятью", "Запускает функцию конкурентно", "Создаёт процесс", "Закрывает канал"},
			CorrectOption: 1,
		},
		{
			Text:          "Какой тег идиоматичен для JSON-сериализации поля?",
			Options:       []string{"`json:Name`", "@json(\"Name\")", "`json:\"name\"`", "[json:name]"},
			CorrectOption: 2,
		},
		{
			Text:          "Что произойдёт при close() уже закрытого канала?",
			Options:       []string{"Игнорируется", "panic", "Канал переоткроется", "Возвращает false"},
			CorrectOption: 1,
		},
		{
			Text:          "defer выполняется...",
			Options:       []string{"перед return", "в момент defer", "в обратном порядке после return", "никогда"},
			CorrectOption: 2,
		},
	},
	"python": {
		{
			Text:          "Какой менеджер пакетов стандартный для Python?",
			Options:       []string{"npm", "pip", "cargo", "gem"},
			CorrectOption: 1,
		},
		{
			Text:          "Чем list отличается от tuple?",
			Options:       []string{"Ничем", "list изменяемый, tuple — нет", "list типизирован, tuple — нет", "tuple быстрее, list безопаснее"},
			CorrectOption: 1,
		},
		{
			Text:          "Что делает декоратор @staticmethod?",
			Options:       []string{"Делает метод неизменяемым", "Превращает в classmethod", "Метод не получает self/cls", "Кэширует результат"},
			CorrectOption: 2,
		},
		{
			Text:          "Какой оператор для эл-ва в коллекции?",
			Options:       []string{"contains", "exists", "in", "of"},
			CorrectOption: 2,
		},
		{
			Text:          "GIL — это...",
			Options:       []string{"библиотека GUI", "Global Interpreter Lock", "GraphQL Integration Layer", "Garbage Indirection Limit"},
			CorrectOption: 1,
		},
	},
	"react": {
		{
			Text:          "Чем useState отличается от useReducer?",
			Options:       []string{"Ничем", "useState только для строк", "useReducer для сложного state с экшенами", "useReducer асинхронный"},
			CorrectOption: 2,
		},
		{
			Text:          "Когда вызывается useEffect без deps?",
			Options:       []string{"Один раз при mount", "На каждый render", "Никогда", "Только на unmount"},
			CorrectOption: 1,
		},
		{
			Text:          "Что такое key prop в списках?",
			Options:       []string{"CSS-селектор", "Уникальный идентификатор для reconciliation", "Имя ref'а", "API-токен"},
			CorrectOption: 1,
		},
		{
			Text:          "React.memo — это про...",
			Options:       []string{"локальный state", "мемоизацию рендера компонента", "кэш fetch-запросов", "управление формами"},
			CorrectOption: 1,
		},
		{
			Text:          "Где правильно делать data-fetching в функциональном компоненте?",
			Options:       []string{"В теле компонента напрямую", "В useEffect", "В render-функции", "В пропсах"},
			CorrectOption: 1,
		},
	},
	"postgresql": {
		{
			Text:          "Что делает EXPLAIN ANALYZE?",
			Options:       []string{"Удаляет таблицу", "Показывает план запроса с реальным временем", "Анализирует индекс", "Открывает транзакцию"},
			CorrectOption: 1,
		},
		{
			Text:          "Какой индекс лучше для LIKE 'abc%'?",
			Options:       []string{"GIN", "B-tree", "BRIN", "Hash"},
			CorrectOption: 1,
		},
		{
			Text:          "Уровень изоляции по умолчанию в PostgreSQL?",
			Options:       []string{"Read Uncommitted", "Read Committed", "Repeatable Read", "Serializable"},
			CorrectOption: 1,
		},
		{
			Text:          "Что значит VACUUM?",
			Options:       []string{"Дроп таблицы", "Резервная копия", "Очистка мёртвых строк MVCC", "Перезагрузка кластера"},
			CorrectOption: 2,
		},
		{
			Text:          "Какой тип лучше для UUID?",
			Options:       []string{"VARCHAR(36)", "TEXT", "UUID", "BIGINT"},
			CorrectOption: 2,
		},
	},
	"docker": {
		{
			Text:          "Что описывает Dockerfile?",
			Options:       []string{"Только сеть", "Шаги сборки образа", "Хранилище секретов", "Композицию сервисов"},
			CorrectOption: 1,
		},
		{
			Text:          "В чём разница между ENTRYPOINT и CMD?",
			Options:       []string{"Это синонимы", "ENTRYPOINT задаёт исполняемый, CMD — аргументы", "CMD только для shell", "ENTRYPOINT не используется в Linux"},
			CorrectOption: 1,
		},
		{
			Text:          "docker-compose нужен для...",
			Options:       []string{"Сборки одного образа", "Описания и запуска многоконтейнерного приложения", "Деплоя на k8s", "Управления секретами"},
			CorrectOption: 1,
		},
		{
			Text:          "Что такое volume?",
			Options:       []string{"Имя образа", "Точка монтирования для персистентных данных", "Сетевой адаптер", "Лимит CPU"},
			CorrectOption: 1,
		},
		{
			Text:          "Какой command строит образ?",
			Options:       []string{"docker pull", "docker run", "docker build", "docker compose"},
			CorrectOption: 2,
		},
	},
	"java": {
		{
			Text:          "Что такое JVM?",
			Options:       []string{"Just Variable Modifier", "Виртуальная машина Java", "Java Virtual Memory", "Jvm Velocity Manager"},
			CorrectOption: 1,
		},
		{
			Text:          "Что делает ключевое слово final у класса?",
			Options:       []string{"Запрещает наследование", "Делает поля immutable", "Помечает класс как abstract", "Финализирует объект"},
			CorrectOption: 0,
		},
		{
			Text:          "Какой коллекцией пользоваться для FIFO?",
			Options:       []string{"HashMap", "ArrayDeque / Queue", "TreeSet", "Stack"},
			CorrectOption: 1,
		},
		{
			Text:          "Что такое Generics?",
			Options:       []string{"Шаблонные типы", "Java-фреймворк", "Аналог reflection", "Альтернатива interfaces"},
			CorrectOption: 0,
		},
		{
			Text:          "checked-исключения проверяются...",
			Options:       []string{"в runtime", "компилятором", "линтером", "JVM при загрузке"},
			CorrectOption: 1,
		},
	},
}

// GetExpertiseTest возвращает вопросы и порог для теста по навыку.
func (s *UsersService) GetExpertiseTest(ctx context.Context, slug string) (*usersv1.ExpertiseTest, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return nil, fmt.Errorf("%w: skill_slug is required", ErrInvalidProfileData)
	}
	qs, ok := expertiseCatalog[slug]
	if !ok {
		return &usersv1.ExpertiseTest{
			SkillSlug: slug,
			Available: false,
			Reason:    "Для этого навыка тест ещё не готов. Эксперт может ревьюить только заявки на сертификаты/курсы.",
		}, nil
	}
	out := &usersv1.ExpertiseTest{
		SkillSlug:        slug,
		Available:        true,
		PassThresholdPct: passThresholdPct,
	}
	for i, q := range qs {
		out.Questions = append(out.Questions, &usersv1.TestQuestion{
			Id:      int32(i),
			Text:    q.Text,
			Options: q.Options,
		})
	}
	return out, nil
}

// SubmitExpertiseTest проверяет ответы. При passed → добавляет slug в expert_verified_skill_slugs.
func (s *UsersService) SubmitExpertiseTest(ctx context.Context, userID, slug string, answers []int32) (*usersv1.SubmitExpertiseTestResponse, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("%w: invalid uuid", ErrInvalidProfileData)
	}
	slug = strings.ToLower(strings.TrimSpace(slug))
	qs, ok := expertiseCatalog[slug]
	if !ok {
		return nil, fmt.Errorf("%w: no test for this skill", ErrInvalidProfileData)
	}
	if len(answers) != len(qs) {
		return nil, fmt.Errorf("%w: answers length mismatch (got %d, want %d)", ErrInvalidProfileData, len(answers), len(qs))
	}
	correct := 0
	for i, a := range answers {
		if int(a) == qs[i].CorrectOption {
			correct++
		}
	}
	total := len(qs)
	scorePct := int32((correct * 100) / total)
	passed := scorePct >= passThresholdPct

	resp := &usersv1.SubmitExpertiseTestResponse{
		Passed:   passed,
		Correct:  int32(correct),
		Total:    int32(total),
		ScorePct: scorePct,
	}
	if passed {
		// Добавляем в expert_verified_skill_slugs union.
		if err := s.repo.Users.AddExpertVerifiedSkills(ctx, userID, []string{slug}); err != nil {
			if errors.Is(err, repository.ErrProfileNotFound) {
				return nil, ErrProfileNotFound
			}
			return nil, fmt.Errorf("failed to record expert-verified skill: %w", err)
		}
		resp.Message = fmt.Sprintf("Тест пройден (%d/%d). Навык #%s добавлен в подтверждённую экспертизу.", correct, total, slug)
	} else {
		resp.Message = fmt.Sprintf("Тест не пройден (%d/%d, нужно ≥%d%%). Попробуйте ещё раз.", correct, total, passThresholdPct)
	}
	return resp, nil
}
