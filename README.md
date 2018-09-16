# 概要
goqueryを使用してgoogle playの開発元をgoogle検索し、検索結果に日本語サイトがあるかどうかで日本法人があるかどうかをざっくり判定する。

## 条件
* 開発元を検索した結果のタイトルに、日本語が入っていれば日本法人がある可能性が高いという前提のもと、
* 検索結果の3つ目まで確認し、Alphabet以外の文字列がタイトルに含まれているかどうかを判定する
* ただし、Alphabet以外の文字列がタイトルに含まれている場合でも、以下の条件に合致する結果はスキップする。
  * リンクURLのドメインにgoogle play、iTunes、facebookが入っている
  * タイトルに"画像検索結果"が入っている
