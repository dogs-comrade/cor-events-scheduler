package services

import (
	"context"
	"math"
	"sort"
	"time"

	"cor-events-scheduler/internal/config"
	"cor-events-scheduler/internal/domain/models"
)

type AnalysisService struct {
	minTechBreak          int
	equipmentComplexity   float64
	multidayBuffer        int
	weatherRiskMultiplier float64
	humanFactorMultiplier float64
	equipmentRiskBase     float64
}

func NewAnalysisService(cfg *config.Config) *AnalysisService {
	return &AnalysisService{
		minTechBreak:          cfg.Analysis.MinTechBreakDuration,
		equipmentComplexity:   cfg.Analysis.EquipmentComplexity,
		multidayBuffer:        cfg.Analysis.MultidayBuffer,
		weatherRiskMultiplier: cfg.Analysis.WeatherRiskMultiplier,
		humanFactorMultiplier: cfg.Analysis.HumanFactorMultiplier,
		equipmentRiskBase:     cfg.Analysis.EquipmentRiskBase,
	}
}

// CalculateTechBreak вычисляет необходимую длительность технического перерыва
func (s *AnalysisService) CalculateTechBreak(block *models.Block, nextBlock *models.Block) int {
	baseBreak := s.minTechBreak

	// Учитываем сложность оборудования
	equipmentFactor := 1.0
	for _, eq := range block.Equipment {
		equipmentFactor += eq.ComplexityScore * s.equipmentComplexity
	}

	// Учитываем количество участников
	participantFactor := math.Log2(float64(block.MaxParticipants)) / 4

	// Учитываем смену локации
	locationChange := 1.0
	if nextBlock != nil && block.Location != nextBlock.Location {
		locationChange = 1.5
	}

	totalBreak := float64(baseBreak) * equipmentFactor * (1 + participantFactor) * locationChange
	return int(math.Ceil(totalBreak))
}

// CalculateScheduleRisk вычисляет вероятность задержек в расписании
func (s *AnalysisService) CalculateScheduleRisk(schedule *models.Schedule) (float64, []string) {
	var totalRisk float64
	var recommendations []string

	// Базовый риск для каждого блока
	blockRisks := make(map[uint]float64)

	for _, block := range schedule.Blocks {
		risk := s.calculateBlockRisk(block)
		blockRisks[block.ID] = risk

		// Анализ зависимостей
		for _, depID := range block.Dependencies {
			if depRisk, exists := blockRisks[depID]; exists {
				risk *= (1 + depRisk/2) // Зависимые блоки увеличивают риск
			}
		}

		// Анализ временных промежутков
		if len(block.Items) > 0 {
			timeGapRisk := s.analyzeTimeGaps(block.Items)
			risk *= (1 + timeGapRisk)
		}

		totalRisk += risk

		// Генерация рекомендаций
		if risk > 0.3 {
			recommendations = append(recommendations, s.generateRecommendations(block, risk))
		}
	}

	// Учитываем многодневность мероприятия
	if isMultiDay := s.isMultiDayEvent(schedule.EventID); isMultiDay {
		totalRisk *= 1.2
		recommendations = append(recommendations,
			"Многодневное мероприятие: рекомендуется добавить дополнительный буфер между днями")
	}

	return totalRisk, recommendations
}

func (s *AnalysisService) calculateBlockRisk(block models.Block) float64 {
	var risk float64

	// Базовый риск на основе сложности блока
	risk = block.Complexity * s.equipmentRiskBase

	// Учет рисков оборудования
	for _, eq := range block.Equipment {
		risk += eq.ComplexityScore * s.equipmentRiskBase
	}

	// Анализ специфических факторов риска
	for _, factor := range block.RiskFactors {
		switch factor.Type {
		case "weather":
			risk += factor.Probability * factor.Impact * s.weatherRiskMultiplier
		case "human":
			risk += factor.Probability * factor.Impact * s.humanFactorMultiplier
		case "equipment":
			risk += factor.Probability * factor.Impact * s.equipmentComplexity
		}
	}

	// Учет количества участников
	participantRisk := math.Log2(float64(block.MaxParticipants)) / 10
	risk += participantRisk

	// Учет требуемого персонала
	staffRisk := float64(block.RequiredStaff) * 0.02
	risk += staffRisk

	return math.Min(risk, 1.0) // Нормализация риска до 1
}

func (s *AnalysisService) analyzeTimeGaps(items []models.BlockItem) float64 {
	if len(items) < 2 {
		return 0
	}

	var totalGapRisk float64
	for i := 0; i < len(items)-1; i++ {
		currentItem := items[i]
		nextItem := items[i+1]

		// Анализ времени между элементами
		gap := nextItem.Duration - currentItem.Duration
		if gap < 5 { // Слишком маленький промежуток
			totalGapRisk += 0.1
		}
	}

	return totalGapRisk
}

func (s *AnalysisService) isMultiDayEvent(eventID uint) bool {
	// Получаем информацию о мероприятии из базы данных
	// Проверяем, является ли оно многодневным
	// Для примера просто вернем true
	return true
}

func (s *AnalysisService) generateRecommendations(block models.Block, risk float64) string {
	var recommendation string

	if risk > 0.7 {
		recommendation = "Критический риск для блока '" + block.Name + "': "
	} else if risk > 0.5 {
		recommendation = "Высокий риск для блока '" + block.Name + "': "
	} else {
		recommendation = "Средний риск для блока '" + block.Name + "': "
	}

	// Анализ конкретных факторов
	if block.MaxParticipants > 20 {
		recommendation += "Рекомендуется разделить на подгруппы. "
	}

	if len(block.Equipment) > 5 {
		recommendation += "Большое количество оборудования - увеличьте тех. перерыв. "
	}

	if block.Complexity > 0.7 {
		recommendation += "Высокая сложность - добавьте дополнительный персонал. "
	}

	if len(block.Dependencies) > 2 {
		recommendation += "Много зависимостей - рассмотрите упрощение структуры. "
	}

	return recommendation
}

// OptimizeSchedule оптимизирует расписание на основе анализа рисков
func (s *AnalysisService) OptimizeSchedule(ctx context.Context, schedule *models.Schedule) (*models.Schedule, error) {
	// Копируем расписание для оптимизации
	optimizedSchedule := *schedule

	// Рассчитываем текущий риск
	currentRisk, _ := s.CalculateScheduleRisk(schedule)

	// Оптимизируем технические перерывы
	s.optimizeTechBreaks(&optimizedSchedule)

	// Добавляем буферное время для блоков с высоким риском
	s.addRiskBuffers(&optimizedSchedule)

	// Перераспределяем блоки, если необходимо
	s.reorderBlocks(&optimizedSchedule)

	// Проверяем, что оптимизация улучшила ситуацию
	newRisk, _ := s.CalculateScheduleRisk(&optimizedSchedule)
	if newRisk >= currentRisk {
		return schedule, nil // Возвращаем оригинальное расписание, если оптимизация не помогла
	}

	return &optimizedSchedule, nil
}

func (s *AnalysisService) optimizeTechBreaks(schedule *models.Schedule) {
	for i := 0; i < len(schedule.Blocks)-1; i++ {
		currentBlock := &schedule.Blocks[i]
		nextBlock := &schedule.Blocks[i+1]

		// Рассчитываем оптимальный технический перерыв
		optimalBreak := s.CalculateTechBreak(currentBlock, nextBlock)

		// Если текущий перерыв меньше оптимального, увеличиваем его
		if currentBlock.TechBreakDuration < optimalBreak {
			currentBlock.TechBreakDuration = optimalBreak

			// Обновляем время начала следующего блока
			nextBlockStart := currentBlock.StartTime.Add(
				time.Duration(currentBlock.Duration+currentBlock.TechBreakDuration) * time.Minute,
			)
			nextBlock.StartTime = nextBlockStart
		}
	}
}

func (s *AnalysisService) addRiskBuffers(schedule *models.Schedule) {
	for i, block := range schedule.Blocks {
		risk := s.calculateBlockRisk(block)

		if risk > 0.5 { // Для блоков с высоким риском
			bufferTime := int(math.Ceil(float64(block.Duration) * 0.2)) // 20% от длительности блока
			schedule.Blocks[i].TechBreakDuration += bufferTime
		}
	}
}

func (s *AnalysisService) reorderBlocks(schedule *models.Schedule) {
	// Сортируем блоки по риску (более рискованные ставим раньше, когда больше времени на решение проблем)
	blocks := schedule.Blocks
	sort.Slice(blocks, func(i, j int) bool {
		riskI := s.calculateBlockRisk(blocks[i])
		riskJ := s.calculateBlockRisk(blocks[j])
		return riskI > riskJ
	})

	// Обновляем порядок и времена начала
	currentTime := blocks[0].StartTime
	for i := range blocks {
		blocks[i].Order = i + 1
		blocks[i].StartTime = currentTime
		currentTime = currentTime.Add(
			time.Duration(blocks[i].Duration+blocks[i].TechBreakDuration) * time.Minute,
		)
	}

	schedule.Blocks = blocks
}
