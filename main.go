package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Question описывает вопрос теста.
type Question struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// Answer описывает ответ пользователя – оценка для вопроса.
type Answer struct {
	QuestionID int `json:"questionId"`
	Value      int `json:"value"` // число: -2, -1, 0, 1, 2
}

// SubmitRequest описывает структуру запроса с ответами пользователя.
type SubmitRequest struct {
	Answers []Answer `json:"answers"`
}

// InterestArea – направление, для которого указаны номера вопросов (бланк ответов).
type InterestArea struct {
	Name        string `json:"name"`
	QuestionIDs []int  `json:"questionIds"`
}

// TestData объединяет инструкцию, список направлений и вопросы.
type TestData struct {
	Instruction   string         `json:"instruction"`
	InterestAreas []InterestArea `json:"interestAreas"`
	Questions     []Question     `json:"questions"`
}

var (
	// Перечень направлений. Эти данные можно оставить статичными или загрузить аналогичным способом.
	interestAreas = []InterestArea{
		{"Биология", []int{1, 30, 59, 88, 117}},
		{"География", []int{2, 31, 60, 89, 118}},
		{"Геология", []int{3, 32, 61, 90, 119}},
		{"Медицина", []int{4, 33, 62, 91, 120}},
		{"Легкая и пищевая промышленность", []int{5, 34, 63, 92, 121}},
		{"Физика", []int{6, 35, 64, 93, 122}},
		{"Химия", []int{7, 36, 65, 94, 123}},
		{"Техника, механика", []int{8, 37, 66, 95, 124}},
		{"Электротехника, радиотехника, электроника", []int{9, 38, 67, 96, 125}},
		{"Обработка материалов (дерево, металл и т.п.)", []int{10, 39, 68, 97, 126}},
		{"Информационные технологии", []int{11, 40, 69, 98, 127}},
		{"Психология", []int{12, 41, 70, 99, 128}},
		{"Строительство", []int{13, 42, 71, 100, 129}},
		{"Транспорт, авиация, морское дело", []int{14, 43, 72, 101, 130}},
		{"Военные специальности", []int{15, 44, 73, 102, 131}},
		{"История", []int{16, 45, 74, 103, 132}},
		{"Литература, филология", []int{17, 46, 75, 104, 133}},
		{"Журналистика, связи с общественностью, реклама", []int{18, 47, 76, 105, 134}},
		{"Социология, философия", []int{19, 48, 77, 106, 135}},
		{"Педагогика", []int{20, 49, 78, 107, 136}},
		{"Право, юриспруденция", []int{21, 50, 79, 108, 137}},
		{"Сфера обслуживания", []int{22, 51, 80, 109, 138}},
		{"Математика", []int{23, 52, 81, 110, 139}},
		{"Экономика, бизнес", []int{24, 53, 82, 111, 140}},
		{"Иностранные языки, лингвистика", []int{25, 54, 83, 112, 141}},
		{"Изобразительное искусство", []int{26, 55, 84, 113, 142}},
		{"Сценическое искусство", []int{27, 56, 85, 114, 143}},
		{"Музыка", []int{28, 57, 86, 115, 144}},
		{"Физкультура, спорт", []int{29, 58, 87, 116, 145}},
	}

	// Текст инструкции для теста.
	instruction = `Инструкция: Вам предстоит оценить свои интересы в пределах 29 направлений.
Выберите один из пяти возможных вариантов. По каждому направлению можно набрать от -10 до 10 баллов.
Набранные от -10 до -5 баллов свидетельствуют о явном отрицании интереса к направлению, -5-0 баллов - отсутствие интереса,
1-3 балла - слабый интерес, 4-6 баллов - средне выраженный интерес, 7-10 баллов - ярко выраженный интерес.`

	// Глобальная переменная для тестовых данных.
	testData TestData
)

// loadQuestions читает вопросы из текстового файла, где каждая строка имеет формат:
// "Номер. Текст вопроса"
func loadQuestions(fileName string) ([]Question, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var questions []Question
	scanner := bufio.NewScanner(file)

	// Регулярное выражение для строки вида: "1. Знакомиться с жизнью растений и животных."
	re := regexp.MustCompile(`^(\d+)\.\s*(.+)$`)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Printf("Ошибка преобразования номера (%s): %v\n", matches[1], err)
				continue
			}
			questionText := matches[2]
			questions = append(questions, Question{ID: id, Text: questionText})
		} else {
			log.Printf("Строка не соответствует формату: %q\n", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return questions, nil
}

// testHandler возвращает тестовые данные (инструкцию, направления и вопросы).
func testHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем кросс-доменные запросы.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(testData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type AreaResult struct {
	Area   string `json:"area"`
	Score  int    `json:"score"`  // можно выводить для отладки
	Result string `json:"result"` // описание интереса, например, «интерес ярко выражен»
}

// getDescription возвращает описание интереса по сумме баллов.
func getDescription(sum int) string {
	if sum >= 7 && sum <= 10 {
		return "интерес ярко выражен"
	} else if sum >= 4 && sum <= 6 {
		return "интерес средне выражен"
	} else if sum >= 1 && sum <= 3 {
		return "слабо выраженный интерес"
	} else if sum >= -4 && sum <= 0 {
		return "интерес отрицается"
	} else if sum >= -10 && sum <= -5 {
		return "интерес явно отрицается"
	}
	return "не определено"
}

// submitHandler принимает ответы пользователя, суммирует оценки по направлениям,
// преобразует их в описание и возвращает отсортированный список.
func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Обработка preflight-запроса
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	// Заголовки для поддержки CORS.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var submitReq SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&submitReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем отображение: ID вопроса → введённая оценка.
	answersMap := make(map[int]int)
	for _, ans := range submitReq.Answers {
		answersMap[ans.QuestionID] = ans.Value
	}

	// Для каждого направления суммируем оценки.
	var areaResults []AreaResult
	for _, area := range interestAreas {
		sum := 0
		for _, qid := range area.QuestionIDs {
			sum += answersMap[qid] // если вопрос не отвечен, значение по умолчанию – 0
		}
		areaResults = append(areaResults, AreaResult{
			Area:   area.Name,
			Score:  sum,
			Result: getDescription(sum),
		})
	}

	// Сортируем направления по убыванию суммы баллов.
	sort.Slice(areaResults, func(i, j int) bool {
		return areaResults[i].Score > areaResults[j].Score
	})

	// Отдаем отсортированный результат.
	if err := json.NewEncoder(w).Encode(areaResults); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Загружаем вопросы из текстового файла (например, "questions.txt").
	loadedQuestions, err := loadQuestions("questions.txt")
	if err != nil {
		log.Fatalf("Ошибка загрузки вопросов: %v\n", err)
	}

	// Собираем тестовые данные, используя загруженные вопросы.
	testData = TestData{
		Instruction:   instruction,
		InterestAreas: interestAreas,
		Questions:     loadedQuestions,
	}

	http.HandleFunc("/api/test", testHandler)
	http.HandleFunc("/api/submit", submitHandler)

	// Добавляем поддержку динамического порта для хостинга
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Значение по умолчанию для локального запуска
	}
	log.Println("Сервер запущен на порту:", port)

	// Запуск сервера
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
