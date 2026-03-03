// Package staticlint предоставляет единый multichecker для статического анализа Go-кода.
//
// Multichecker построен на базе фреймворка golang.org/x/tools/go/analysis
// и объединяет три группы анализаторов:
//
//  1. Стандартные анализаторы из пакета golang.org/x/tools/go/analysis/passes,
//     по сути — те же проверки, что используются в go vet.
//  2. Все анализаторы класса SA, а также остальные проверки из набора staticcheck
//     (пакет honnef.co/go/tools/staticcheck).
//  3. Пользовательский анализатор checker.Analyzer из локального пакета
//     github.com/sirajDeveloper/metrics-alerts-collector/cmd/staticlint/checker.
//
// # Механизм запуска multichecker
//
// Multichecker работает как обычный консольный инструмент. После сборки бинаря
// вы можете передать ему пути к пакетам, как это делается с go test или go vet:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Под капотом используется функция multichecker.Main из пакета
// golang.org/x/tools/go/analysis/multichecker. Она получает список *analysis.Analyzer,
// далее:
//
//   - строит список целевых пакетов (по аргументам командной строки);
//
//   - для каждого пакета:
//
//   - парсит исходный код в AST;
//
//   - строит типовую информацию (types.Info, types.Package);
//
//   - запускает все анализаторы, учитывая их зависимости (Requires);
//
//   - собирает и выводит диагностики (ошибки, предупреждения, подсказки).
//
// Каждый анализатор представляет собой структуру analysis.Analyzer, в которой
// описаны:
//
//   - Name: имя анализатора (используется в выводе и для конфигурации);
//   - Doc: краткое описание назначения;
//   - Run: функция, выполняющая анализ (получает *analysis.Pass и может вызывать
//     pass.Reportf для генерации диагностик);
//   - Requires: зависимые анализаторы (если нужно предварительно собрать типы,
//     операции SSA и т.п.).
//
// Таким образом, staticlint даёт единый вход для запуска множества проверок
// за один проход по коду.
//
// # Группы анализаторов и их назначение
//
// ### Анализаторы staticcheck (honnef.co/go/tools/staticcheck)
//
// В блоке:
//
//	for _, v := range staticcheck.Analyzers {
//	    analyzers = append(analyzers, v.Analyzer)
//	}
//
// подключаются все публичные анализаторы из набора staticcheck, включая:
//
//   - SAxxxx — «serious issues», потенциальные баги (nil deref, misuse of sync,
//     неправильная работа с каналами и т.д.).
//   - Sxxxx, STxxxx, QFxxxx и другие классы — стиль, упрощения кода,
//     потенциальные улучшения, подсказки, quickfix-предложения.
//
// Staticcheck считается «расширенным go vet»: он находит многие логические,
// конкурентные и API-ошибки, которые не покрываются стандартными анализаторами.
//
// Документация по каждому конкретному анализатору доступна по адресу
// https://staticcheck.dev/docs/checks.
//
// ### Стандартные анализаторы golang.org/x/tools/go/analysis/passes
//
// Ниже перечислены стандартные анализаторы, которые явным образом добавляются
// в multichecker, и их назначение:
//
//   - asmdecl.Analyzer
//     Проверяет соответствие объявлений функций в Go и их реализаций на asm.
//     Помогает поймать несоответствие сигнатур при использовании ассемблера.
//
//   - assign.Analyzer
//     Ищет подозрительные присваивания, такие как x = x, а также ошибки
//     с присваиванием в многомерных выражениях.
//
//   - atomic.Analyzer
//     Проверяет корректность использования пакета sync/atomic: обнаруживает
//     некорректные типы операндов, небезопасные операции над полями структур.
//
//   - bools.Analyzer
//     Находит странные логические выражения (например, if x == true),
//     избыточные или всегда истинные/ложные условия, упрощаемые конструкции.
//
//   - buildtag.Analyzer
//     Проверяет корректность build tags в исходниках (формат, расположение,
//     совместимость с правилами компилятора).
//
//   - cgocall.Analyzer
//     Ищет нарушения правил использования cgo: опасные передачи указателей
//     в C-код, проблемы с жизненным циклом объектов и т.п.
//
//   - copylock.Analyzer
//     Предупреждает о копировании структур, содержащих sync.Mutex, RWMutex
//     и другие небезопасно копируемые типы. Помогает избежать тонких
//     гонок и deadlock-ов.
//
//   - deepequalerrors.Analyzer
//     Находит некорректное использование reflect.DeepEqual с типом error,
//     когда сравнение по значению может быть некорректным.
//
//   - errorsas.Analyzer
//     Проверяет корректность использования функции errors.As, следит за тем,
//     чтобы целевые аргументы имели подходящий тип.
//
//   - httpresponse.Analyzer
//     Ищет утечки http.Response.Body и некорректное закрытие тела ответа,
//     что важно для сетевых клиентов и серверов.
//
//   - ifaceassert.Analyzer
//     Находит потенциально опасные утверждения типов (type assertion) над
//     интерфейсами, которые могут паниковать или не учитывают полный набор типов.
//
//   - loopclosure.Analyzer
//     Классическая проверка на захват переменной цикла в замыкание или горутину:
//     предупреждает о ситуациях, когда все горутины используют одно и то же
//     значение переменной цикла.
//
//   - lostcancel.Analyzer
//     Проверяет, что функции context.WithCancel/WithTimeout/WithDeadline
//     сопровождаются корректным вызовом cancel(), чтобы избежать утечек ресурсов.
//
//   - nilfunc.Analyzer
//     Находит вызовы nil-функций, когда переменная-функция может быть nil,
//     что приведёт к панике во время выполнения.
//
//   - nilness.Analyzer
//     Анализирует возможную nil-ность указателей и ссылочных типов, выявляя
//     потенциальные разыменования nil.
//
//   - printf.Analyzer
//     Проверяет корректность форматных строк для функций семейства fmt.Printf:
//     соответствие количества/типа аргументов спецификаторам формата.
//
//   - shadow.Analyzer
//     Ищет затенение (shadowing) переменных во внутренних областях видимости,
//     когда новая переменная перекрывает внешнюю, что часто приводит
//     к логическим ошибкам.
//
//   - shift.Analyzer
//     Находит подозрительные операции битового сдвига (например, отрицательный
//     сдвиг, сдвиг на величину, превышающую размер типа).
//
//   - stdmethods.Analyzer
//     Проверяет методы, которые по имени похожи на стандартные (например, Error,
//     Len, Less), но имеют некорректную сигнатуру и потому могут не вызываться
//     ожидаемым образом.
//
//   - stringintconv.Analyzer
//     Предупреждает о странных преобразованиях int↔string, например, string(65),
//     когда на самом деле подразумевалась конвертация числа в десятичную строку.
//
//   - structtag.Analyzer
//     Валидирует теги структур (json, xml, yaml и т.п.): формат, дубли,
//     некорректные ключи.
//
//   - tests.Analyzer
//     Ищет типичные ошибки в тестах: неправильные имена тестовых функций,
//     сигнатуры, использование testing.T и т.д.
//
//   - timeformat.Analyzer
//     Проверяет строки формата времени для time.Parse/Format, ищет
//     некорректные шаблоны.
//
//   - unmarshal.Analyzer
//     Анализирует вызовы json.Unmarshal/xml.Unmarshal и подобные, выявляя
//     некорректные типы целей и другие типичные ошибки десериализации.
//
//   - unreachable.Analyzer
//     Находит недостижимый код (например, после return/panic), который
//     может указывать на логические ошибки.
//
//   - unsafeptr.Analyzer
//     Предупреждает об опасных операциях с unsafe.Pointer, нарушающих
//     правила памяти и выравнивания.
//
//   - unusedresult.Analyzer
//     Проверяет, не проигнорирован ли важный возвращаемый результат функций,
//     где его нужно всегда обрабатывать (например, regexp.Compile).
//
//   - unusedwrite.Analyzer
//     Ищет записи в переменные/поля/элементы, значения которых потом
//     ни разу не читаются, что может указывать на лишний или ошибочный код.
//
// ### Пользовательский анализатор checker.Analyzer
//
// В конце списка анализаторов добавляется:
//
//	analyzers = append(analyzers, checker.Analyzer)
//
// Анализатор checker.Analyzer определён в локальном пакете
// github.com/sirajDeveloper/metrics-alerts-collector/cmd/staticlint/checker
// и реализует специфичные для проекта правила. Его назначение зависит
// от реализации checker.Analyzer, но типично это:
//
//   - проверка внутренних код-стайлов;
//   - запрет использования отдельных API или паттернов (например, прямой os.Exit
//     в main, прямой логгер вместо обёртки и т.п.);
//   - доменно-специфичные инварианты (валидация метрик, имени алертов и др.).
//
// # Использование
//
// Подключите пакет staticlint в свой cmd-пакет или экспортируйте main-функцию
// напрямую (как в примере ниже), соберите бинарь и запускайте его на любом
// Go-проекте:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Все описанные выше анализаторы будут выполнены для каждого пакета,
// а найденные проблемы будут выведены в стандартный вывод в формате,
// совместимом с привычными инструментами Go.
package checker

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"

	"honnef.co/go/tools/staticcheck"
)

// Main запускает multichecker со всеми подключёнными анализаторами.
//
// Обычно эту функцию вызывают из пакета cmd/staticlint:
//
//	package main
//
//	import "github.com/sirajDeveloper/metrics-alerts-collector/cmd/staticlint"
//
//	func main() {
//	    staticlint.Main()
//	}
func Main() {
	var analyzers []*analysis.Analyzer

	// Анализаторы staticcheck (включая весь класс SA).
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes.
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
	)

	// Пользовательский анализатор проекта.
	analyzers = append(analyzers, Analyzer)

	multichecker.Main(analyzers...)
}
