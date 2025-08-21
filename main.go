package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"slices"
	"strings"
)

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ComposeImages 将 files 二维数组里的图片拼接成一张大图。
// 每幅图放在 (128 * 行号, 128 * 列号) 处，空字符串表示该位置留空。
// 返回拼接后的 *image.RGBA。
func ComposeImages(files [][]string, path string) (*image.RGBA, error) {
	// 计算画布大小
	rows := len(files)
	if rows == 0 {
		return nil, fmt.Errorf("empty files array")
	}
	cols := 0
	for _, row := range files {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return nil, fmt.Errorf("empty files array")
	}

	// 最终画布尺寸
	width := cols * 128
	height := rows * 128

	// 创建画布
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	// 逐格贴图
	for r, row := range files {
		for c, name := range row {
			if name == "" {
				continue
			}

			f, err := os.Open(path + name + ".png")
			if err != nil {
				return nil, fmt.Errorf("open %s: %w", name, err)
			}
			img, err := png.Decode(f)
			f.Close()
			if err != nil {
				return nil, fmt.Errorf("decode %s: %w", name, err)
			}

			// 将 img 绘制到目标画布
			bounds := img.Bounds()
			pos := image.Pt(c*128, r*128)
			draw.Draw(dst,
				image.Rectangle{Min: pos, Max: pos.Add(bounds.Size())},
				img,
				bounds.Min,
				draw.Src)
		}
	}
	return dst, nil
}

func GenImage(files [][]string, filename string, path string) {
	out, err := ComposeImages(files, path)
	if err != nil {
		log.Fatal(err)
	}
	// 保存到 out.png
	f, err := os.Create("./output/" + filename + ".png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		log.Fatal(err)
	}
	// fmt.Println("saved to out.png")
}

func StationStatus(stations []string, current string, ltr bool) []string {
	out := make([]string, len(stations))
	nCurrent := slices.Index(stations, current)
	for i := range stations {
		if (i == 0 && ltr) || (i == len(stations)-1 && !ltr) {
			// 起点
			if i == nCurrent {
				// 当前
				if !ltr {
					// 向左
					out[i] = "@"
				} else {
					// 向右
					out[i] = "@#"
				}
			} else {
				if !ltr {
					out[i] = "="
				} else {
					out[i] = "#="
				}
			}
		} else if (i == 0 && !ltr) || (i == len(stations)-1 && ltr) {
			//终点
			if i == nCurrent {
				//当前
				if !ltr {
					// 向左
					out[i] = "@="
				} else {
					// 向右
					out[i] = "@#="
				}
			} else {
				if !ltr {
					out[i] = ""
				} else {
					out[i] = "#"
				}
			}
		} else {
			// 中途
			if i == nCurrent {
				//当前
				if !ltr {
					out[i] = "@"
				} else {
					out[i] = "@="
				}
			} else if (i < nCurrent && !ltr) || (i > nCurrent && ltr) {
				//未到
				out[i] = ""
			} else {
				out[i] = "="
			}
		}
	}
	return out
}

func Arrange(stations, center []string, current string, ltr bool, length int, blank string) ([]string, error) {
	switch len(center) {
	case 1:
		if length%2 == 0 {
			return nil, errors.New("单元素居中时 length 必须为奇数")
		}
	case 2:
		// 不做奇偶要求
	default:
		return nil, errors.New("center 长度只能为 1 或 2")
	}

	if len(stations) > length {
		return nil, errors.New("stations 长度不能大于 length")
	}

	// 1. 找到 center 在 stations 中的起始下标 p
	var p int
	switch length % 2 {
	case 1:
		for i, v := range stations {
			if v == center[0] {
				p = i
				goto found
			}
		}
		return nil, errors.New("center 单元素不在 stations 中")
	case 0:
		for i := 0; i <= len(stations)-2; i++ {
			if stations[i] == center[0] && stations[i+1] == center[1] {
				p = i
				goto found
			}
		}
		return nil, errors.New("center 两元素不是 stations 中相邻且顺序一致")
	}

found:
	var centerStartWant int
	switch len(center) {
	case 1:
		centerStartWant = length / 2
	case 2:
		centerStartWant = length/2 - 1 // 双元素中心起点
	}
	offset := centerStartWant - p

	// 4. 构建结果
	out := make([]string, length)
	for i := range out {
		out[i] = blank
	}
	status := StationStatus(stations, current, ltr)
	for i := range stations {
		out[offset+i] = "tc" + stations[i] + status[i]
	}
	return out, nil
}

// func Arrange(stations []string, center []string, current string, ltr bool, length int, blank string) ([]string, error) {
// 	if len(stations) > length {
// 		return make([]string, length), errors.New("stations 长度不能大于 14")
// 	}
// 	if len(center) != 2 {
// 		return make([]string, length), errors.New("center 必须恰好 2 个元素")
// 	}

// 	// 验证 center 是否真的是 stations 中相邻且顺序一致的两个元素
// 	found := -1
// 	for i := 0; i <= len(stations)-2; i++ {
// 		if stations[i] == center[0] && stations[i+1] == center[1] {
// 			found = i
// 			break
// 		}
// 	}
// 	if found == -1 {
// 		return make([]string, length), errors.New("center 不是 stations 中相邻且顺序一致的两个元素")
// 	}
// 	status := StationStatus(stations, current, ltr)
// 	out := make([]string, length)
// 	for i := range out {
// 		out[i] = blank
// 	}
// 	L := len(stations)
// 	start := (length - L) / 2 // 居中起始下标
// 	for i := range stations {
// 		out[start+i] = "tc" + stations[i] + status[i]
// 	}
// 	return out, nil
// }

func ComposeStations(stations []string, center []string, config [][]string, path string) {
	line := stations[0][:len(stations[0])-3]
	fmt.Println("正在生成" + line + "号线……")
	for num, station := range stations {
		if config[0][num][0] == '<' {
			// 加向，向左
			direction := "+"
			arrow := ""
			towards := []string{"tczhongdianzhan", ""}
			if num != 0 {
				arrow = "tczuo"
				towards = []string{"tckaiwang", "tckw" + line + direction}
			}
			if config[0][num] == "<2" {
				row1 := []string{
					"", "", "", "", "", "",
					arrow,
					"tc" + line + direction,
					"tcfenge",
					"tczm" + station, "",
					towards[0], towards[1], "",
					"", "", "", "", "", "",
				}
				row2, _ := Arrange(stations, center, station, false, 20, "tc"+line)
				GenImage([][]string{row1, row2}, "sh"+station+"+", path)
			} else {
				row1 := []string{
					arrow,
					"tc" + line + direction,
					towards[0], towards[1], "",
					"tczm" + station, "",
				}
				length := 15
				if len(stations) <= 14 {
					row1 = append(row1, "tc"+line+"zuo")
					length = 14
				}
				row2, _ := Arrange(stations, center, station, false, length, "tc"+line)
				row := [][]string{append(row1, row2...)}
				GenImage(row, "dh"+station+"+", path)
			}
		}
		if config[1][num][0] == '>' {
			// 减向，向右
			direction := "-"
			arrow := ""
			towards := []string{"", "tczhongdianzhan", ""}
			if num != len(stations)-1 {
				arrow = "tcyou"
				towards = []string{"tckaiwang", "tckw" + line + direction, ""}
			}
			if config[1][num] == ">2" {
				row1 := []string{
					"", "", "", "", "", "",
					towards[0], towards[1], towards[2],
					"tczm" + station, "",
					"tcfenge",
					"tc" + line + direction,
					arrow,
					"", "", "", "", "", "",
				}
				row2, _ := Arrange(stations, center, station, true, 20, "tc"+line)
				GenImage([][]string{row1, row2}, "sh"+station+"-#", path)
			} else {
				row1 := []string{
					"tczm" + station, "",
					towards[0], towards[1], towards[2],
					"tc" + line + direction,
					arrow,
				}
				length := 15
				if len(stations) <= 14 {
					row1 = append([]string{"tc" + line + "you"}, row1...)
					length = 14
				}
				row2, _ := Arrange(stations, center, station, true, length, "tc"+line)
				row := [][]string{append(row2, row1...)}
				GenImage(row, "dh"+station+"-#", path)
			}
		}
	}
	// sort.Slice(stations, func(i, j int) bool { return i > j })
	// sort.Slice(center, func(i, j int) bool { return i > j })
	reverse(stations)
	reverse(center)
	reverse(config[0])
	reverse(config[1])
	for num, station := range stations {
		if config[1][num][0] == '<' {
			// 减向，向左
			direction := "-"
			arrow := ""
			towards := []string{"tczhongdianzhan", ""}
			if num != 0 {
				arrow = "tczuo"
				towards = []string{"tckaiwang", "tckw" + line + direction}
			}
			if config[1][num] == "<2" {
				row1 := []string{
					"", "", "", "", "", "",
					arrow,
					"tc" + line + direction,
					"tcfenge",
					"tczm" + station, "",
					towards[0], towards[1], "",
					"", "", "", "", "", "",
				}
				row2, _ := Arrange(stations, center, station, false, 20, "tc"+line)
				GenImage([][]string{row1, row2}, "sh"+station+"-", path)
			} else {
				row1 := []string{
					arrow,
					"tc" + line + direction,
					towards[0], towards[1], "",
					"tczm" + station, "",
				}
				length := 15
				if len(stations) <= 14 {
					row1 = append(row1, "tc"+line+"zuo")
					length = 14
				}
				row2, _ := Arrange(stations, center, station, false, length, "tc"+line)
				row := [][]string{append(row1, row2...)}
				GenImage(row, "dh"+station+"-", path)
			}
		}
		if config[0][num][0] == '>' {
			// 加向，向右
			direction := "+"
			arrow := ""
			towards := []string{"", "tczhongdianzhan", ""}
			if num != len(stations)-1 {
				arrow = "tcyou"
				towards = []string{"tckaiwang", "tckw" + line + direction, ""}
			}
			if config[0][num] == ">2" {
				row1 := []string{
					"", "", "", "", "", "",
					towards[0], towards[1], towards[2],
					"tczm" + station, "",
					"tcfenge",
					"tc" + line + direction,
					arrow,
					"", "", "", "", "", "",
				}
				row2, _ := Arrange(stations, center, station, true, 20, "tc"+line)
				GenImage([][]string{row1, row2}, "sh"+station+"+#", path)
			} else {
				row1 := []string{
					"tczm" + station, "",
					towards[0], towards[1], towards[2],
					"tc" + line + direction,
					arrow,
				}
				length := 15
				if len(stations) <= 14 {
					row1 = append([]string{"tc" + line + "you"}, row1...)
					length = 14
				}
				row2, _ := Arrange(stations, center, station, true, length, "tc"+line)
				row := [][]string{append(row2, row1...)}
				GenImage(row, "dh"+station+"+#", path)
			}
		}
	}
}

// 一段数据
type segment struct {
	stations []string
	center   []string
	config   [][]string
}

func ReadStationsFile(path string) ([]segment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var segs []segment
	scanner := bufio.NewScanner(file)

	// 读段
	for {
		seg, ok := readSegment(scanner)
		if !ok {
			break
		}
		segs = append(segs, seg)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return segs, nil
}

// 读一段
func readSegment(scanner *bufio.Scanner) (seg segment, ok bool) {
	var (
		stations []string
		center   []string
		row0     []string
		row1     []string
		// seen     = make(map[string]bool)
		first = true
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break // 空行结束本段
		}
		line = strings.ReplaceAll(line, " ", "")

		sep := strings.LastIndex(line, ":")
		if sep == -1 {
			sep = strings.Index(line, ",")
		}
		if sep == -1 {
			continue
		}
		station := line[:sep]
		right := strings.Split(line[sep+1:], ",")

		// 去重保持顺序
		// if !seen[station] {
		// 	seen[station] = true
		// 	stations = append(stations, station)
		// }

		// 首行定义 center
		if first {
			first = false
			if len(right) >= 2 {
				center = []string{right[0], right[1]}
			} else if len(right) == 1 {
				center = []string{right[0]}
			}
		} else {
			stations = append(stations, station)
			// 填充两行
			switch len(right) {
			case 1:
				row0 = append(row0, right[0])
				row1 = append(row1, right[0])
			case 2:
				row0 = append(row0, right[0])
				row1 = append(row1, right[1])
			}
		}
	}

	if len(stations) == 0 {
		return segment{}, false // 无数据
	}
	return segment{stations: stations, center: center, config: [][]string{row0, row1}}, true
}

func main() {
	fmt.Println("正在读取stations.txt……")
	segs, err := ReadStationsFile("stations.txt")
	if err != nil {
		panic(err)
	}
	for _, seg := range segs {
		ComposeStations(seg.stations, seg.center, seg.config, "./images/")
	}
}

// func main() {
// 	//dt1
// 	stations := []string{"1+06", "1+05", "1+04", "1+03", "1+02", "1+01", "1+00", "1-01", "1-02", "1-03", "1--", "1--", "1-06"}
// 	center := []string{"1+01", "1+00"}
// 	config1 := []string{">1", ">1", ">1", "<2", "<1", "<1", ">1", "<2", "<2", "<2", "0", "0", ">2"}
// 	config2 := make([]string, len(config1))
// 	copy(config2, config1)
// 	config := [][]string{config1, config2}
// 	ComposeStations(stations, center, config)
// 	//dt2
// 	stations = []string{"2+07", "2--", "2--", "2+04", "2+03", "2+02", "2+01", "2+00", "2-01", "2-02", "2-03", "2--", "2-06", "2-07"}
// 	center = []string{"2+01", "2+00"}
// 	config1 = []string{"<1", "0", "0", "<1", "<2", "<2", "<1", "<1", "<1", "<2", "<2", "0", "<2", "<2"}
// 	config2 = make([]string, len(config1))
// 	copy(config2, config1)
// 	config = [][]string{config1, config2}
// 	ComposeStations(stations, center, config)
// 	//dt3
// 	stations = []string{"3+04", "3+03", "3+02", "3+01", "3+00", "3-01", "3-02"}
// 	center = []string{"3+01", "3+00"}
// 	config1 = []string{"<1", "<2", "<1", "<1", "<1", "0", "<1"}
// 	config2 = make([]string, len(config1))
// 	copy(config2, config1)
// 	config = [][]string{config1, config2}
// 	ComposeStations(stations, center, config)
// 	//dt4
// 	stations = []string{"4+03", "4+02", "4+01", "4+00"}
// 	center = []string{"4+01", "4+00"}
// 	config1 = []string{">1", ">1", ">1", "<2"}
// 	config2 = make([]string, len(config1))
// 	copy(config2, config1)
// 	config = [][]string{config1, config2}
// 	ComposeStations(stations, center, config)
// 	//dt8
// 	stations = []string{"8+02", "8+01", "8+00", "8-01", "8-02"}
// 	center = []string{"8+01", "8+00"}
// 	config1 = []string{"<2", "<2", "<2", "<2", "<2"}
// 	config2 = []string{"<2", ">2", "<2", "<2", "<2"}
// 	config = [][]string{config1, config2}
// 	ComposeStations(stations, center, config)
// 	// 1) 打开文件
// 	// fileName := "D:/MC/papersoc/plugins/Denizen/data/创游集.yml"
// 	// f, err := os.Open(fileName)
// 	// if err != nil {
// 	// 	log.Fatalf("打开文件失败: %v", err)
// 	// }
// 	// defer f.Close()

// 	// // 2) 解析到 map（无需结构体）
// 	// var data map[string]any
// 	// if err := yaml.NewDecoder(f).Decode(&data); err != nil {
// 	// 	log.Fatalf("解析 YAML 失败: %v", err)
// 	// }
// 	// if node, ok := data["轨交"].(map[string]any); ok {
// 	// 	for k, v := range node {
// 	// 		if k[0:2] == "kt" {
// 	// 			stations := v.(map[string]any)["车站"].(map[string]any)["站号"]
// 	// 		}
// 	// 	}
// 	// }
// }
