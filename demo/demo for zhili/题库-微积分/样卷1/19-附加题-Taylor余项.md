---
app: requiz
bank: 题库-微积分
id: WJ019
path: 样卷1/19-附加题-Taylor余项.md
chapter: 微积分A1 期末考试样卷
grade: 大学一年级
difficulty: ★★★★
importance: 选做挑战（A+）
source: 微积分A1期末考试样卷
knowledge: Taylor展式与余项
type: 证明题
---

## 题目

设 $f(x)$ 为闭区间 $[-h, h]$ 上无穷可微函数，其中 $h > 0$。假设对 $\forall x \in [0, h]$，以及任意的非负整数 $n$，都有 $f^{(n)}(x) \ge 0$。记

$$r_n(x) = \frac{1}{n!}\int_0^x (x-t)^n f^{(n+1)}(t)\,dt$$

求证：$\forall x \in (0, h)$，均有 $\lim_{n\to+\infty}r_n(x) = 0$。

## 答案

结论成立，证明如下。

## 解析

注意 $r_n(x)$ 是函数 $f(x)$ 在点 $x = 0$ 处 $n$ 阶 Taylor 展式的 Cauchy 型积分余项，即

$$f(x) = \sum_{k=0}^{n}\frac{f^{(k)}(0)}{k!}x^k + r_n(x) \qquad (*)$$

对积分 $\int_0^x (x-t)^n f^{(n+1)}(t)\,dt$ 作变量代换 $x - t = xu$，则

$$r_n(x) = \frac{1}{n!}\int_0^1 (xu)^n f^{(n+1)}\bigl(x(1-u)\bigr)\cdot x\,du = \frac{x^{n+1}}{n!}\int_0^1 u^n f^{(n+1)}\bigl(x(1-u)\bigr)\,du$$

上式可写作 $\frac{r_n(x)}{x^{n+1}} = \frac{1}{n!}\int_0^1 u^n f^{(n+1)}\bigl(x(1-u)\bigr)\,du$。

根据假设函数 $f(x)$ 的各阶导数非负，可知（因 $0 < x < h$ 且 $1 - u \ge 0$）

$$f^{(n+1)}\bigl(x(1-u)\bigr) \le f^{(n+1)}\bigl(h(1-u)\bigr)$$

因此 $\frac{r_n(x)}{x^{n+1}} \le \frac{r_n(h)}{h^{n+1}}$，$x \in (0, h)$。

再由展式 $(*)$ 及各阶导数非负可知 $r_n(h) \le f(h)$。于是对于任意 $x \in (0, h)$：

$$0 \le r_n(x) \le \frac{x^{n+1}}{h^{n+1}}r_n(h) = \left(\frac{x}{h}\right)^{n+1}f(h) \to 0 \qquad (n \to +\infty)$$

命题得证。

## 备注

附加题全对才给分，分数不计入总评，仅用于评判 A+。本题关键：非负导数假设 + 单调性把 $x$ 处的余项放大到 $h$ 处，再借助 Taylor 展式控制 $r_n(h)$。
