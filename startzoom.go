package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Config struct {
	Classes    []ClassData `json:"Classes"`
	SumId      int         `json:"SumId"`
	TimeMargin int         `json:"TimeMargin"`
	IsAsk      bool        `json:"IsAsk"`
}

/*授業の情報を格納する構造体*/
type ClassData struct {
	Id 		int `json:"Id"`
	Name    string `json:"Name"`
	Weekday string `json:"Weekday"`
	Date 	string `json:"Date"`
	Start   string `json:"Start"`
	End     string `json:"End"`
	Url     string `json:"Url"`
}

/*Zoomデータ構造体が空かどうか返す関数*/
func (cd ClassData) isEmpty() bool {
	if cd.Name == "" {
		return true
	}
	return false
}

var sc = bufio.NewScanner(os.Stdin)

/*入力読み込み用関数*/
func read() string {
	sc.Scan()
	return sc.Text()
}

/*数値入力用関数*/
func InputNum (msg string) int {
	for {
		fmt.Println(msg)
		i, e := strconv.Atoi(read())
		if e != nil {
			continue
		}
		return i
	}
}

/*ファイルの存在を確認する関数*/
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

/*jsonファイルを読み込んで構造体を返す関数*/
func loadClasses(filename string) (config Config) {
	if !fileExists(filename) {
		if _, err := os.Create(filename); err != nil {
			log.Fatal(err)
		}
		config.SumId = 0
		config.TimeMargin = 10
		config.IsAsk = false
	}
	bytes, err := ioutil.ReadFile(filename)	//json読み込み
	if err != nil {
		log.Fatal(err)
	}
	if len(bytes) != 0 {
		if err := json.Unmarshal(bytes, &config); err != nil {
			log.Fatal(err)
		}
	}
	return
}

/*jsonファイルに書き込む関数*/
func saveConfig(config Config, filename string) {
	classJson, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	fp, err := os.OpenFile(filename, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	_, err = fp.Write(classJson)
	if err != nil {
		panic(err)
	}
}

/*授業の名前を入力する関数*/
func inputName() (name string) {
	fmt.Print("\n授業名を入力:")
	name = read()
	return
}

/*授業の曜日を入力する関数*/
func inputWeekday() (weekday string, date string) {
	date = ""
	weekday = ""
	fmt.Println("\nZoomが開催される曜日を指定します また、毎週開催されるものでなくある日程のみのZoomの場合は、日付のみの指定も可能です")
	switch InputNum("曜日(または日付指定)を選択: 0: 日付で指定する, 1: Sunday, 2: Monday, 3: Tuesday, 4: Wednesday, 5: Thursday, 6: Friday, 7: Saturday") {
	case 0:
		tmp := InputNum("日付を入力(例：1月2日 => 0102 (半角数字))")
		date = strconv.Itoa(tmp / 100) + "-" + strconv.Itoa(tmp % 100)
	case 1: weekday = "Sunday"
	case 2: weekday = "Monday"
	case 3: weekday = "Tuesday"
	case 4: weekday = "Wednesday"
	case 5: weekday = "Thursday"
	case 6: weekday = "Friday"
	case 7: weekday = "Saturday"
	default: weekday, date = inputWeekday()
	}
	return
}

/*授業の開始時刻を入力する関数*/
func inputStartTime() (startTime string) {
	fmt.Print("\n")
	tmp := InputNum("開始時刻を入力(例：14:30 => 1430 (半角数字))")
	startTime = strconv.Itoa(tmp / 100) + ":" + strconv.Itoa(tmp % 100)
	if tmp % 100 == 0 { startTime += "0" }
	return
}

/*授業の終了時刻を入力する関数*/
func inputEndTime() (endTime string) {
	fmt.Print("\n")
	tmp := InputNum("終了時刻を入力")
	endTime = strconv.Itoa(tmp / 100) + ":" + strconv.Itoa(tmp % 100)
	if tmp % 100 == 0 { endTime += "0" }
	return
}

/*授業のURLを入力する関数*/
func inputUrl() (url string) {
	fmt.Print("\nZoomURLを入力:")
	url = read()
	return
}

/*新規登録する授業の構造体を作成する関数*/
func makeClass(id int) (cd ClassData) {
	cd.Id = id
	cd.Name = inputName()
	cd.Weekday, cd.Date = inputWeekday()
	cd.Start = inputStartTime()
	cd.End = inputEndTime()
	cd.Url = inputUrl()
	fmt.Println(cd.Name, "を作成しました")
	return
}

/*ZoomデータをもとにURLを開く関数*/
func runZoom(cd ClassData)  {
	fmt.Println(cd.Name, "のZoomを開きます")
	err := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", cd.Url).Start()
	if err != nil {
		panic(err)
	}
}

/*Zoomデータから起動する時刻かどうか調べる関数*/
func checkTime(cd ClassData, timeMargin int) bool {
	now := time.Now()
	nowTime, _ := time.Parse("15:4", strconv.Itoa(now.Hour())+ ":" + strconv.Itoa(now.Minute()))
	startTime, _ := time.Parse("15:04", cd.Start)
	startTime = startTime.Add(time.Duration(-1 * timeMargin) * time.Minute)
	endTime, _ := time.Parse("15:04", cd.End)
	if startTime.Before(nowTime) && endTime.After(nowTime) {
		return true
	}
	return false
}

/*現在時刻より遅いかつ開始の早い方のZoomデータを返す関数*/
func getEarlierClass(data1 ClassData, data2 ClassData) ClassData {
	now := time.Now()
	nowTime, _ := time.Parse("15:4", strconv.Itoa(now.Hour())+ ":" + strconv.Itoa(now.Minute()))
	time1, _ := time.Parse("15:04", data1.Start)
	time2, _ := time.Parse("15:04", data2.Start)
	if nowTime.After(time1) && nowTime.After(time2) {
		var cd ClassData
		return cd
	} else if nowTime.After(time1) {
		return data2
	} else if nowTime.After(time2) {
		return data1
	} else {
		if time1.Before(time2) {
			return data1
		}
		return data2
	}
}

/*曜日か日付が合致するZoomを探す関数*/
func startZoom(config Config) {
	classes := config.Classes
	var nextClass ClassData
	var currentClass ClassData
	now := time.Now()
	hour := strconv.Itoa(now.Hour())
	min := strconv.Itoa(now.Minute())
	if now.Minute() < 10 {
		min = "0" + min
	}
	fmt.Println("現在時刻:", hour, ":", min)
	_, month, day := now.Date()
	today := strconv.Itoa(int(month)) + "-" + strconv.Itoa(day)
	for _, cd := range classes {
		if cd.Date == today {
			if checkTime(cd, config.TimeMargin) {
				currentClass = cd
			}
			if nextClass.isEmpty() {
				nextClass = cd
			} else {
				nextClass = getEarlierClass(nextClass, cd)
			}
		}
	}
	for _, cd := range classes {
		if cd.Weekday == now.Weekday().String() {
			if checkTime(cd, config.TimeMargin) {
				currentClass = cd
			}
			if nextClass.isEmpty() {
				nextClass = cd
			} else {
				nextClass = getEarlierClass(nextClass, cd)
			}
		}
	}
	if currentClass.isEmpty() {
		fmt.Println("現在または", config.TimeMargin, "分後に進行中の授業はありません")
		fmt.Print("\n")
		if !nextClass.isEmpty() && config.IsAsk {
			msg := nextClass.Start + " から " + nextClass.Name + " が始まりますが、起動しますか？" +
				"\n1: はい, 2: いいえ"
			if InputNum(msg) == 1 {
				runZoom(nextClass)
			} else {
				fmt.Println("起動せず戻ります")
			}
		}
	} else {
		if !nextClass.isEmpty() && checkTime(nextClass, config.TimeMargin) {
			runZoom(nextClass)
		} else {
			runZoom(currentClass)
		}
	}
}

/*授業単体の情報を表示する関数*/
func showClassData(cd ClassData) {
	fmt.Println(cd.Name)
	var dayOrDate string
	if cd.Date == "" {
		dayOrDate = cd.Weekday
	} else {
		dayOrDate = cd.Date
	}
	fmt.Println("", dayOrDate, cd.Start, "~", cd.End)
	fmt.Println("", cd.Url)
}

/*登録授業のリストを表示する関数*/
func showClassList(classes []ClassData) {
	fmt.Println("\n登録されている授業を表示します.")
	fmt.Print("\n")
	if len(classes) == 0 {
		fmt.Println("登録授業なし")
	} else {
		for i, cd := range classes {
			fmt.Print("\n", i+1, ": ")
			showClassData(cd)
		}
	}
}

/*登録授業単体を編集する関数*/
func editClassData(cd ClassData) (editedCd ClassData) {
	editedCd = cd
	switch InputNum(editedCd.Name + "の何を編集しますか？\n" +
					"1: 名前, 2: 曜日または日付, 3: 開始時刻, 4: 終了時刻, 5: URL, 6: すべて") {
	case 1: editedCd.Name = inputName()
	case 2: editedCd.Weekday, editedCd.Date = inputWeekday()
	case 3: editedCd.Start = inputStartTime()
	case 4: editedCd.End = inputEndTime()
	case 5: editedCd.Url = inputUrl()
	case 6:
		fmt.Println("すべて編集します")
		editedCd = makeClass(editedCd.Id)
	default:
		editedCd = editClassData(cd)
	}
	return editedCd
}

/*登録授業リストを編集する関数*/
func editClasses(classes []ClassData) (editedClasses []ClassData) {
	fmt.Println("\n登録授業の編集をします")
	showClassList(classes)
	fmt.Print("\n")
	classNum := InputNum("編集したい授業の番号を入力してください(編集せず戻る場合は「0」)")
	if classNum == 0 {
		fmt.Println("編集せずに戻ります")
		return classes
	} else {
		classNum -= 1
		if classNum >= len(classes) || classNum < 0 {
			fmt.Println("授業の番号が不正です")
			return classes
		} else {
			editedClasses = classes
			editedClasses[classNum] = editClassData(classes[classNum])
			fmt.Println("\n編集が正常に終了しました")
			fmt.Print("\n")
			showClassData(editedClasses[classNum])
		}
	}
	return
}

/*登録授業単体の削除を行う関数*/
func deleteClassData(classes []ClassData, index int) (editedClasses []ClassData) {
	for i, cd := range classes {
		if i == index { continue }
		editedClasses = append(editedClasses, cd)
	}
	return
}

/*登録授業の削除を行う関数*/
func deleteClasses(classes []ClassData) (editedClasses []ClassData) {
	fmt.Println("\n登録授業の削除をします")
	showClassList(classes)
	fmt.Print("\n")
	classNum := InputNum("削除したい授業の番号を入力してください(すべて削除する場合は「-1」)(削除せず戻る場合は「0」)")
	if classNum == 0 {
		fmt.Println("削除せずに戻ります")
		return classes
	} else if classNum == -1 {
		fmt.Println("すべての授業データを削除します よろしいですか？")
		switch InputNum("1: はい, 2: いいえ") {
		case 1:
			fmt.Println("すべてのデータを削除しました")
			return editedClasses
		default:
			fmt.Println("削除せずに戻ります")
			return classes
		}
	} else {
		classNum -= 1
		if classNum >= len(classes) || classNum < 0 {
			fmt.Println("授業の番号が不正です")
			return classes
		} else {
			fmt.Println(classes[classNum].Name, "の授業データを削除します よろしいですか？")
			switch InputNum("1: はい, 2: いいえ") {
			case 1:
				fmt.Println(classes[classNum].Name, "のデータを削除します)")
				editedClasses = classes
				editedClasses = deleteClassData(classes, classNum)
				fmt.Println("\n削除が正常に終了しました")
			case 2:
				fmt.Println("削除せずに戻ります")
				return classes
			}
		}
	}
	return
}

/*登録授業を編集・削除する関数*/
func editDeleteClasses(classes []ClassData) (editedClasses []ClassData) {
	fmt.Println("\n登録授業の編集・削除を行います")
	if len(classes) == 0 {
		fmt.Println("登録授業なし")
		return classes
	}
	switch InputNum("0: 戻る, 1: 編集, 2: 削除") {
	case 1: editedClasses = editClasses(classes)
	case 2: editedClasses = deleteClasses(classes)
	default: return classes
	}
	return
}

/*選択してZoomを開始する関数*/
func anytimeStart(classes []ClassData) {
	fmt.Println("\nZoom選んでを起動します")
	showClassList(classes)
	fmt.Print("\n")
	classNum := InputNum("起動するZoomの番号を入力(戻る場合は0)")
	if classNum == 0 {
		fmt.Println("戻ります")
		return
	}
	classNum--
	if classNum >= len(classes) || classNum < 0 {
		fmt.Println("授業の番号が不正です")
		return
	}
	runZoom(classes[classNum])
}

/*開始前の時間の余裕を設定する関数*/
func editTimeMargin(config Config) (timeMargin int) {
	fmt.Println("\nZoom開始時刻の何分前から起動するようにするか設定します(現在は", config.TimeMargin, "分)")
	return InputNum("何分前から起動可能に設定しますか？")
}

/*該当Zoomがないときに近いZoomを開くかどうかを設定する関数*/
func editIsAsk() bool {
	fmt.Println("授業開始を選択した際に、開始時刻に該当するZoomがなかったときに、同じ日のなかで" +
					"最も開始時刻の近いZoomを開くかどうかの質問の有無を設定します")
	if InputNum("1: 聞く, 2: 聞かない") == 1 {
		return true
	} else {
		return false
	}
}

/*設定変更を行う関数*/
func editConfig(config Config) (editedConfig Config) {
	editedConfig = config
	fmt.Println("\n設定の変更をします")
	switch InputNum("0: 戻る, 1: Zoom開始前の余裕時間, 2: 該当Zoomがない場合の質問") {
	case 1: editedConfig.TimeMargin = editTimeMargin(config)
	case 2: editedConfig.IsAsk = editIsAsk()
	default: return config
	}
	fmt.Println("設定を変更しました")
	return
}

/*メイン関数*/
func StartZoomMain(opts Options) {
	filename := "config.json"
	config := loadClasses(filename)

	if len(opts.Start) != 0 {
		startZoom(config)
		return
	}

	fmt.Println("\n" +
		"---------------------------------------------\n" +
		"----------------- StartZoom -----------------\n" +
		"----------- (made by RikuTsuzuki) -----------\n" +
		"---------------------------------------------")
	fmt.Print("\n")

	flg := 0
	for flg == 0 {
		switch InputNum("\n行いたい操作の番号を入力してください\n0: 終了, 1: 授業開始, 2: 授業登録, 3: 授業リスト, 4: 登録授業の編集・削除, 5: 選択して授業開始, 6: 設定") {
		case 0:
			fmt.Println("終了します")
			flg = 1
		case 1:
			startZoom(config)
		case 2:
			fmt.Println("新しく授業を登録します。")
			config.SumId++
			config.Classes = append(config.Classes, makeClass(config.SumId))
			saveConfig(config, filename)
		case 3:
			showClassList(config.Classes)
		case 4:
			config.Classes = editDeleteClasses(config.Classes)
			saveConfig(config, filename)
		case 5:
			anytimeStart(config.Classes)
		case 6:
			config = editConfig(config)
			saveConfig(config, filename)
		default:
		}
	}
}
