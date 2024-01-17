package main

import (
	"log"
	"os/exec"
)

func main() {
	out, err := exec.Command("diff-pdf", "-mv", "--output-diff=diff.pdf", "theory_1.pdf", "theory_2.pdf").CombinedOutput()
	// diff-pdf コマンドは実効性工事でも exit status 1 を返すので、
	// exit status 1 以外のエラーが発生した場合のみエラーとして扱う
	if err != nil && err.Error() != "exit status 1" {
		log.Fatalln("diff-pdf コマンドの実行に失敗しました。: ", err)
	}
	log.Println(string(out))
}
