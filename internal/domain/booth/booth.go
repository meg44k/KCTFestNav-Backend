// クラスの展示、部活の出店に関するプログラム

package booth

import "errors"

// クラス展示、部活の出店などをあらわす型。
// 混雑度は専用メソッドから読み取り、書き込みを行う必要があります。
type Booth struct {
	Name             string // ブース名
	Organizer        string // ブースの主催者(ex. 1-1, 陸上部...)
	Detail           string // ブースの説明
	congestionStatus int    // 0: 空き 1: 少し混雑している 2: かなり混雑している
}

// 新しいBoothのインスタンスを作成します
func NewBooth(
	name string,
	organizer string,
	detail string,
) *Booth {
	return &Booth{congestionStatus: 0}
}

func (b *Booth) CongestionStatus() int {
	return b.congestionStatus
}

func (b *Booth) SetCongestionStatus(congestionLevel int) error {
	if 0 > congestionLevel || congestionLevel > 2 {
		return errors.New("Congestion level must be between 0 and 2")
	}
	b.congestionStatus = congestionLevel
	return nil
}
