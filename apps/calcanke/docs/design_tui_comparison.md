# TUIライブラリ比較

## 候補ライブラリ

### 1. bubbletea + bubbles（推奨）

**Charm社のモダンなTUIフレームワーク**

- **pros**:
  - Elmアーキテクチャ（モデル・ビュー・更新）で保守性が高い
  - bubblesで豊富なコンポーネント（リスト、テーブル、入力など）
  - Charm社の他ツール（lipgloss、glamourなど）との統合
  - 活発な開発、豊富なサンプル
  - 本格的なTUIアプリに向いている

- **cons**:
  - 学習コスト高め（Elmパターンの理解が必要）
  - シンプルな対話には少しオーバーヘッド

- **使用例**: gh（GitHub CLI）、gum（Charm社のシェルスクリプト用ツール）

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/table"
)
```

### 2. survey

**シンプルな対話型プロンプトライブラリ**

- **pros**:
  - 非常にシンプル、学習コスト低い
  - すぐに使える（10行で動く）
  - 一般的なプロンプト（Select、MultiSelect、Input等）をサポート
  - 軽量

- **cons**:
  - 機能が限定的（複雑なUIには不向き）
  - カスタマイズ性が低い
  - メンテナンスがやや停滞気味

```go
import "github.com/AlecAivazis/survey/v2"

var selected string
prompt := &survey.Select{
    Message: "列を選択:",
    Options: columns,
}
survey.AskOne(prompt, &selected)
```

### 3. tview

**本格的なTUIアプリケーションフレームワーク**

- **pros**:
  - 豊富なウィジェット（テーブル、フォーム、モーダルなど）
  - レイアウトエンジンが強力
  - 複雑なUIも構築可能

- **cons**:
  - 学習コスト高い
  - シンプルな対話には複雑すぎる
  - イベント駆動モデルが独特

### 4. promptui（非推奨）

- シンプルだが、メンテナンスが停止
- surveyの方が優れている

## 推奨の組み合わせ

### パターンA: bubbletea + bubbles（将来性重視）

将来的にWeb UIも作る予定なので、しっかりした設計がベスト。

```go
// アーキテクチャ
type model struct {
    screen      string           // 現在の画面
    columns     []Column         // 列一覧
    list        list.Model       // 列選択リスト
    table       table.Model      // 結果表示テーブル
    selectedX   string           // 選択されたX軸
    selectedY   string           // 選択されたY軸
}

func (m model) Init() tea.Cmd
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m model) View() string
```

**メリット**:
- 画面遷移が管理しやすい
- 状態管理が明確
- 拡張性が高い（将来的に機能追加しやすい）

### パターンB: survey（シンプル重視）

とにかく早く動くものを作りたい場合。

```go
func runInteractive() {
    // 1. 分析タイプ選択
    var analysisType string
    survey.AskOne(&survey.Select{
        Message: "分析タイプ:",
        Options: []string{"単純集計", "クロス集計", ...},
    }, &analysisType)

    // 2. X軸選択
    var xColumn string
    survey.AskOne(&survey.Select{
        Message: "X軸の列:",
        Options: getColumns(),
    }, &xColumn)

    // 3. Y軸選択
    // 4. 集計実行
}
```

**メリット**:
- 実装が簡単
- すぐに動く
- 理解しやすい

## 推奨決定

**段階的なアプローチ**:

### フェーズ1: survey でプロトタイプ（1-2日）
- 最小機能で動作確認
- ユーザーフィードバック取得
- 使い勝手を検証

### フェーズ2: bubbletea に移行（必要に応じて）
- 機能が増えてきたら
- より洗練されたUIが必要になったら
- 画面遷移が複雑になったら

この方が、早く動くものを作りつつ、将来の拡張性も確保できる。

## 次のステップ

1. survey でプロトタイプを実装
2. 使ってみてフィードバック
3. 必要に応じて bubbletea に移行検討
