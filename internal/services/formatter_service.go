package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type FormatterService struct {
	scheduleService *SchedulerService
}

func NewFormatterService(scheduleService *SchedulerService) *FormatterService {
	return &FormatterService{
		scheduleService: scheduleService,
	}
}

func (s *FormatterService) FormatPublicSchedule(ctx context.Context, scheduleID uint) (*models.PublicSchedule, error) {
	schedule, err := s.scheduleService.GetSchedule(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("расписание с ID %d не найдено", scheduleID)
		}
		return nil, fmt.Errorf("ошибка при получении расписания: %w", err)
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

	// Обрабатываем каждый блок расписания
	currentTime := schedule.StartDate
	for _, block := range schedule.Blocks {
		// Создаем основной блок
		scheduleItem := models.PublicScheduleItem{
			Time:        currentTime,
			Title:       block.Name,
			Description: block.Type,
			SubItems:    make([]models.PublicScheduleSubItem, 0),
		}

		// Добавляем подэлементы блока
		subItemTime := currentTime
		for _, item := range block.Items {
			subItem := models.PublicScheduleSubItem{
				Time:        subItemTime,
				Title:       item.Name,
				Description: item.Description,
			}
			scheduleItem.SubItems = append(scheduleItem.SubItems, subItem)
			subItemTime = subItemTime.Add(time.Duration(item.Duration) * time.Minute)
		}

		publicSchedule.Items = append(publicSchedule.Items, scheduleItem)

		// Обновляем время для следующего блока
		currentTime = currentTime.Add(time.Duration(block.Duration) * time.Minute)
		if block.TechBreakDuration > 0 {
			currentTime = currentTime.Add(time.Duration(block.TechBreakDuration) * time.Minute)
		}
	}

	// Сортируем все элементы по времени
	sort.Slice(publicSchedule.Items, func(i, j int) bool {
		return publicSchedule.Items[i].Time.Before(publicSchedule.Items[j].Time)
	})

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

	// Используем локальное время для форматирования
	location, _ := time.LoadLocation("Europe/Moscow")

	for _, item := range schedule.Items {
		// Конвертируем время в локальное
		localTime := item.Time.In(location)
		result.WriteString(fmt.Sprintf("%s %s\n", localTime.Format("15:04"), item.Title))

		// Подэлементы
		for _, subItem := range item.SubItems {
			localSubTime := subItem.Time.In(location)
			result.WriteString(fmt.Sprintf("* %s %s\n", localSubTime.Format("15:04"), subItem.Title))
			if subItem.Description != "" {
				result.WriteString(fmt.Sprintf("  %s\n", subItem.Description))
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
