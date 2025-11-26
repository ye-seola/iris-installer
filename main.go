package main

import (
	"errors"
	"fmt"
	"io"
	"iris_installer/api"
	"iris_installer/mhttp"
	"iris_installer/ps"
	"os"
	"os/exec"
)

const (
	apkPath         = "/data/local/tmp/Iris.apk"
	irisProcessName = "dolidolih_iris"
)

func main() {
	if os.Getuid() != 0 {
		fmt.Println("루트 권한으로 실행해 주세요.")
		return
	}

	cmdName := "./iris_installer"
	if len(os.Args) >= 1 {
		cmdName = os.Args[0]
	}

	if len(os.Args) < 2 {
		printHelp(cmdName)
		return
	}

	switch os.Args[1] {
	case "start":
		handleStart(false)
	case "start-fg":
		handleStart(true)
	case "stop":
		handleStop()
	case "update":
		if err := install(); err != nil {
			panic(err)
		}
	case "check":
		handleCheck()
	case "help":
		printHelp(cmdName)
	default:
		fmt.Printf("알 수 없는 명령어: %s\n", os.Args[1])
	}
}

func printHelp(name string) {
	fmt.Printf("도움말\n")
	fmt.Printf("%s start - Iris를 백그라운드로 시작합니다\n", name)
	fmt.Printf("%s start-fg - Iris를 포그라운드에서 시작합니다\n", name)
	fmt.Printf("%s stop - Iris를 종료합니다\n", name)
	fmt.Printf("%s check - Iris가 실행 중인지 체크합니다\n", name)
	fmt.Printf("%s update - Iris를 최신버전으로 업데이트 합니다\n", name)
}

func handleStart(foreground bool) {
	if !checkAPKExists() {
		fmt.Println("Iris가 설치되어 있지 않습니다. 설치를 시작합니다...")
		if err := install(); err != nil {
			panic(err)
		}
	}

	if err := killIris(); err != nil {
		panic(fmt.Errorf("iris를 종료하지 못했습니다: %w", err))
	}

	fmt.Println("Iris를 시작합니다...")
	if err := startProcess(foreground); err != nil {
		panic(fmt.Errorf("iris를 시작하지 못했습니다: %w", err))
	}
	fmt.Println("Iris 시작 완료.")
}

func handleStop() {
	fmt.Println("Iris 종료 중...")
	if err := killIris(); err != nil {
		panic(fmt.Errorf("iris를 종료하지 못했습니다: %w", err))
	}
	fmt.Println("Iris 종료 완료.")
}

func checkAPKExists() bool {
	_, err := os.Stat(apkPath)
	return err == nil
}

func startProcess(foreground bool) error {
	cmd := exec.Command("app_process", "-cp", apkPath, "/", "--nice-name="+irisProcessName, "party.qwer.iris.Main")
	if foreground {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return cmd.Start()
}

func killIris() error {
	processList, err := ps.GetProcessList()
	if err != nil {
		return err
	}

	for _, p := range processList {
		if p.Name == irisProcessName {
			process, err := os.FindProcess(p.PID)
			if err != nil {
				fmt.Println(fmt.Errorf("프로세스 종료 실패: %w", err))
				continue
			}

			if err := process.Kill(); err != nil {
				fmt.Println(fmt.Errorf("프로세스 종료 실패 %w", err))
				continue
			}

			fmt.Printf("Iris 프로세스(%d)를 종료했습니다.\n", p.PID)
		}
	}

	return nil
}

func install() error {
	info, err := api.GetLatestInfo()
	if err != nil {
		return fmt.Errorf("최신 정보를 가져오지 못했습니다: %w", err)
	}

	fmt.Printf("%s 설치를 시작합니다...\n", info.Name)

	apkData, err := downloadAPK(info.IrisAPKURL)
	if err != nil {
		return err
	}

	if err := verifyDigest(apkData, info.Digest); err != nil {
		return err
	}

	_ = os.Remove(apkPath)
	if err := os.WriteFile(apkPath, apkData, 0555); err != nil {
		return fmt.Errorf("APK 파일 저장 실패: %w", err)
	}

	fmt.Println("설치 완료")
	return nil
}

func downloadAPK(url string) ([]byte, error) {
	httpClient := mhttp.CreateHTTPClient()

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("APK 다운로드 실패: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("APK 읽기 실패: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("다운로드한 APK 데이터가 비어있습니다")
	}

	return data, nil
}

func verifyDigest(data []byte, digest string) error {
	ok, err := api.CheckDigest(data, digest)
	if err != nil {
		return fmt.Errorf("다이제스트 확인 중 오류 발생: %w", err)
	}
	if !ok {
		return errors.New("APK 다이제스트가 일치하지 않습니다")
	}
	return nil
}

func handleCheck() {
	processList, err := ps.GetProcessList()
	if err != nil {
		panic(err)
	}

	for _, p := range processList {
		if p.Name == irisProcessName {
			fmt.Printf("Iris가 실행 중입니다 (%d)\n", p.PID)
		}
	}
}
