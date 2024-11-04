package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/domain/repositories"

	"go.uber.org/zap"
)

type FormatterService struct {
	scheduleService *SchedulerService
	scheduleRepo    *repositories.ScheduleRepository
	logger          *zap.Logger
}

func NewFormatterService(
	schedulerService *SchedulerService,
	scheduleRepo *repositories.ScheduleRepository,
	logger *zap.Logger,
) *FormatterService {
	return &FormatterService{
		scheduleService: schedulerService,
		scheduleRepo:    scheduleRepo,
		logger:          logger,
	}
}

func (s *FormatterService) FormatPublicSchedule(ctx context.Context, scheduleID uint) (*models.PublicSchedule, error) {
	schedule, err := s.scheduleRepo.GetByID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	publicSchedule := &models.PublicSchedule{
		EventName: schedule.Name,
		Date:      schedule.StartDate,
		Items:     make([]models.PublicScheduleItem, 0),
	}

	// Добавляем начало мероприятия
	publicSchedule.Items = append(publicSchedule.Items, models.PublicScheduleItem{
		Time:  schedule.StartDate,
		Title: "Начало мероприятия",
	})

	// Сортируем блоки по порядку
	sort.Slice(schedule.Blocks, func(i, j int) bool {
		return schedule.Blocks[i].Order < schedule.Blocks[j].Order
	})

	currentTime := schedule.StartDate
	for _, block := range schedule.Blocks {
		// Создаем элемент расписания для блока
		blockItem := models.PublicScheduleItem{
			Time:        currentTime,
			Title:       block.Name,
			Description: fmt.Sprintf("%s - %s", block.Type, block.Location),
			SubItems:    make([]models.PublicScheduleSubItem, 0),
		}

		// Сортируем элементы блока
		sort.Slice(block.Items, func(i, j int) bool {
			return block.Items[i].Order < block.Items[j].Order
		})

		// Добавляем подэлементы
		itemTime := currentTime
		for _, item := range block.Items {
			subItem := models.PublicScheduleSubItem{
				Time:        itemTime,
				Title:       item.Name,
				Description: item.Description,
			}

			if len(item.Participants) > 0 {
				var participants []string
				for _, p := range item.Participants {
					participants = append(participants, fmt.Sprintf("%s (%s)", p.Name, p.Role))
				}
				subItem.Description = fmt.Sprintf("%s - %s", subItem.Description, strings.Join(participants, ", "))
			}

			blockItem.SubItems = append(blockItem.SubItems, subItem)
			itemTime = itemTime.Add(time.Duration(item.Duration) * time.Minute)
		}

		publicSchedule.Items = append(publicSchedule.Items, blockItem)

		// Обновляем время для следующего блока
		currentTime = currentTime.Add(time.Duration(block.Duration) * time.Minute)

		// Добавляем технический перерыв, если он есть
		if block.TechBreakDuration > 0 {
			publicSchedule.Items = append(publicSchedule.Items, models.PublicScheduleItem{
				Time:        currentTime,
				Title:       fmt.Sprintf("Технический перерыв (%d мин)", block.TechBreakDuration),
				Description: fmt.Sprintf("После %s", block.Name),
			})
			currentTime = currentTime.Add(time.Duration(block.TechBreakDuration) * time.Minute)
		}
	}

	return publicSchedule, nil
}

// Вспомогательная функция для форматирования времени
func formatEventTime(t time.Time) string {
	return t.Format("15:04")
}

func (s *FormatterService) FormatPublicScheduleText(ctx context.Context, scheduleID uint) (string, error) {
	schedule, err := s.FormatPublicSchedule(ctx, scheduleID)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Расписание: %s\n", schedule.EventName))
	result.WriteString(fmt.Sprintf("Дата: %s\n\n", schedule.Date.Format("02.01.2006")))

	location, _ := time.LoadLocation("Europe/Moscow")

	for _, item := range schedule.Items {
		localTime := item.Time.In(location)
		result.WriteString(fmt.Sprintf("%s %s\n", localTime.Format("15:04"), item.Title))
		if item.Description != "" {
			result.WriteString(fmt.Sprintf("    %s\n", item.Description))
		}

		for _, subItem := range item.SubItems {
			localSubTime := subItem.Time.In(location)
			result.WriteString(fmt.Sprintf("* %s %s\n", localSubTime.Format("15:04"), subItem.Title))
			if subItem.Description != "" {
				result.WriteString(fmt.Sprintf("    %s\n", subItem.Description))
			}
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

func (s *FormatterService) FormatVolunteerSchedule(ctx context.Context, scheduleID uint) (*models.VolunteerSchedule, error) {
	schedule, err := s.scheduleService.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	volunteerSchedule := &models.VolunteerSchedule{
		EventName: schedule.Name,
		Date:      schedule.StartDate,
		Items:     make([]models.VolunteerScheduleItem, 0),
		Notes:     make([]string, 0),
	}

	// Добавляем информацию о подготовке к мероприятию
	setupTime := schedule.StartDate.Add(-1 * time.Hour) // За час до начала
	volunteerSchedule.Items = append(volunteerSchedule.Items, models.VolunteerScheduleItem{
		Time:          setupTime,
		Title:         "Подготовка к мероприятию",
		Location:      "Главный вход",
		Instructions:  "Регистрация волонтеров, получение бейджей и инструктаж",
		RequiredStaff: 5,
	})

	for _, block := range schedule.Blocks {
		// Формируем список оборудования
		equipment := make([]string, 0)
		for _, eq := range block.Equipment {
			equipment = append(equipment, eq.Name)
		}

		// Добавляем информацию о подготовке блока
		setupStart := block.StartTime.Add(-30 * time.Minute)
		volunteerSchedule.Items = append(volunteerSchedule.Items, models.VolunteerScheduleItem{
			Time:          setupStart,
			Title:         fmt.Sprintf("Подготовка к блоку: %s", block.Name),
			Location:      block.Location,
			Equipment:     equipment,
			Instructions:  fmt.Sprintf("Подготовка оборудования и проверка готовности для %s", block.Name),
			RequiredStaff: block.RequiredStaff,
			SetupNotes:    generateSetupNotes(block),
		})

		// Добавляем сам блок
		volunteerSchedule.Items = append(volunteerSchedule.Items, models.VolunteerScheduleItem{
			Time:          block.StartTime,
			Title:         block.Name,
			Location:      block.Location,
			Equipment:     equipment,
			Instructions:  generateInstructions(block),
			RequiredStaff: block.RequiredStaff,
		})

		// Добавляем технический перерыв, если он есть
		if block.TechBreakDuration > 0 {
			breakStart := block.StartTime.Add(time.Duration(block.Duration) * time.Minute)
			volunteerSchedule.Items = append(volunteerSchedule.Items, models.VolunteerScheduleItem{
				Time:          breakStart,
				Title:         "Технический перерыв",
				Location:      block.Location,
				TechBreak:     true,
				BreakDuration: block.TechBreakDuration,
				Instructions:  generateBreakInstructions(block),
				RequiredStaff: block.RequiredStaff / 2, // Обычно на перерыв нужно меньше персонала
			})
		}
	}

	// Сортируем элементы по времени
	sort.Slice(volunteerSchedule.Items, func(i, j int) bool {
		return volunteerSchedule.Items[i].Time.Before(volunteerSchedule.Items[j].Time)
	})

	// Добавляем общие заметки
	volunteerSchedule.Notes = []string{
		"Всегда носите бейдж и форму волонтера",
		"При возникновении проблем обращайтесь к координатору",
		"Следите за временем технических перерывов",
		"Соблюдайте технику безопасности при работе с оборудованием",
	}

	return volunteerSchedule, nil
}

func generateSetupNotes(block models.Block) string {
	notes := []string{
		fmt.Sprintf("Проверить наличие всего оборудования для %s", block.Name),
		"Убедиться в работоспособности технических средств",
	}

	if len(block.Equipment) > 0 {
		notes = append(notes, "Особое внимание уделить:")
		for _, eq := range block.Equipment {
			notes = append(notes, fmt.Sprintf("- %s (время установки: %d мин)", eq.Name, eq.SetupTime))
		}
	}

	return strings.Join(notes, "\n")
}

func generateInstructions(block models.Block) string {
	instructions := []string{
		fmt.Sprintf("Основной блок: %s", block.Name),
		fmt.Sprintf("Максимальное количество участников: %d", block.MaxParticipants),
	}

	if len(block.RiskFactors) > 0 {
		instructions = append(instructions, "\nОсобые указания:")
		for _, risk := range block.RiskFactors {
			instructions = append(instructions, fmt.Sprintf("- %s: %s", risk.Type, risk.Mitigation))
		}
	}

	return strings.Join(instructions, "\n")
}

func generateBreakInstructions(block models.Block) string {
	return fmt.Sprintf(
		"Технический перерыв %d минут\n- Проверить и подготовить оборудование для следующего блока\n- Произвести необходимые перестановки\n- Доложить координатору о готовности",
		block.TechBreakDuration,
	)
}
