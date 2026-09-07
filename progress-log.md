# 📓 Progress Log

> Kunlik ish jurnali — nima o'rganildi, nima yozildi, keyingi qadam nima.
> *Daily work log — what was learned, what was written, what's next.*

🏠 [Repo bosh sahifa](./README.md) · 🧵 [Goroutine](./02-concurrency/goroutine/README.md) · 🟡 [`_Grunnable`](./02-concurrency/goroutine/Grunnable.md)

---

## 📈 Umumiy holat / Overview

| Bo'lim | Hujjat | Kod | Test | Izoh |
|:-------|:------:|:---:|:----:|:-----|
| 01 · Fundamentals | ⬜ | ⬜ | ⬜ | `main.go` skeleti bor |
| 02 · Concurrency  | 🟡 | 🟡 | ⬜ | goroutine 6 ta holati ✅ tugadi; channel va kod hali skelet |
| 03 · DSA          | ⬜ | ⬜ | ⬜ | — |
| 04 · System Design| ⬜ | ⬜ | ⬜ | — |
| 05 · Why-answers  | ⬜ | ⬜ | ⬜ | — |

`⬜ boshlanmagan` · `🟡 jarayonda` · `✅ tugallangan`

---

## 🗓 Jurnal / Log

### 2026-09-07 — Goroutine hayot sikli to'liq yozildi 🎉

Qolgan 4 ta holat uchun alohida hujjat yaratildi — endi **6 ta asosiy holat** to'liq yopilgan.

| Holat | Fayl | Asosiy mavzular |
|:------|:-----|:----------------|
| 🟢 `_Grunning` | [`Grunning.md`](./02-concurrency/goroutine/Grunning.md) | G+M+P uchligi, 3 xil preemption, `sysmon`, `for {}` qotib qolish muammosi |
| 🟠 `_Gsyscall` | [`Gsyscall.md`](./02-concurrency/goroutine/Gsyscall.md) | `entersyscall`/`exitsyscall`, P handoff, tez vs sekin syscall, netpoller |
| 🔵 `_Gwaiting` | [`Gwaiting.md`](./02-concurrency/goroutine/Gwaiting.md) | `gopark`/`goready`, deadlock xatosi, goroutine leak + `context` yechimi |
| ⚫ `_Gdead` | [`Gdead.md`](./02-concurrency/goroutine/Gdead.md) | `gfree` qayta ishlatish, goroutine'ni o'ldirib bo'lmasligi, panic/recover |

**🔗 Bog'lanish:** barcha hujjatlarda ⬅️ ortga / ➡️ keyingi havolalari bor, zanjir:
`README(#gidle)` → `Grunnable` → `Grunning` → `Gsyscall` → `Gwaiting` → `Gdead` → `README`

**📚 Yangi o'rganilgan tushunchalar**

- `_Grunning` goroutine'lar soni hech qachon `GOMAXPROCS`dan oshmaydi
- 3 xil preemption: kooperativ (funksiya chaqiruvida), asinxron (`SIGURG`, Go 1.14+), ixtiyoriy (`Gosched()`)
- `sysmon` — P'siz ishlaydigan nazoratchi thread; 10 ms preempt, 20 µs syscall retake
- Syscall'da **P bo'shatiladi**, G esa M'ga bog'langan qoladi → dastur qotib qolmaydi
- Shu sababli M'lar soni `GOMAXPROCS`dan ko'p bo'ladi (chegara 10 000)
- ⚠️ **Tarmoq I/O `_Gsyscall` emas, `_Gwaiting`** — netpoller (epoll/kqueue/IOCP) ishlatiladi
- `_Grunnable` navbatda + CPU kutadi; `_Gwaiting` navbatda emas + hodisa kutadi
- `all goroutines are asleep` faqat **hamma** goroutine `_Gwaiting` bo'lgandagina chiqadi
- Goroutine leak: abadiy `_Gwaiting` → hech qachon `_Gdead` bo'lmaydi → stek GC'ga tushmaydi
- `_Gdead` → `gfree` (P'da 64 ta lokal + global `sched.gFree`) → `newproc()` qayta ishlatadi
- Goroutine'ni tashqaridan o'ldirish **mumkin emas** — faqat `context`/channel orqali so'rash

---

### 2026-09-07 — Hujjatlar qayta formatlandi

| Nima qilindi | Fayl |
|:-------------|:-----|
| Goroutine hujjati to'liq qayta yozildi: EN/UZ jadvallar, GMP chizmasi, holatlar ro'yxati, `_Gidle` lifecycle | [`02-concurrency/goroutine/README.md`](./02-concurrency/goroutine/README.md) |
| `_Grunnable` hujjati qayta yozildi: state machine, LRQ/GRQ/runnext, work stealing chizmalari | [`02-concurrency/goroutine/Grunnable.md`](./02-concurrency/goroutine/Grunnable.md) |
| Ikki hujjat o'zaro navigatsiya havolalari bilan bog'landi | ↔️ |
| Bosh `README.md` to'ldirildi: bo'limlar jadvali, struktura, ishga tushirish, o'qish tartibi | [`README.md`](./README.md) |
| Ushbu progress log tuzildi | [`progress-log.md`](./progress-log.md) |

**🐞 Tuzatilgan xato / Fixed error**

| | |
|:--|:--|
| **Nima edi** | Eski `README.md`da GMP modelida **M** va **P** ta'riflari almashtirib yuborilgan edi — P "actual operating-system thread" deb yozilgandi. |
| **To'g'risi** | **M** = Machine = **OS thread**. **P** = Processor = **mantiqiy scheduling konteksti** (lokal run queue egasi, soni `GOMAXPROCS`). |
| **Nega muhim** | Bu intervyuda eng ko'p so'raladigan savollardan biri; noto'g'ri o'rganilsa `GOMAXPROCS` mantiqi ham noto'g'ri tushuniladi. |

**📚 O'rganilgan tushunchalar**

- `newproc()` → `_Gidle` → initialized → `_Grunnable` zanjiri
- `gfree` ro'yxati — tugagan `g` struct'lar qayta ishlatiladi
- LRQ 256 ta G sig'imi, to'lganda yarmi GRQ'ga ko'chadi
- `runnext` — 1 o'rinli tezkor slot (kesh yaqinligi uchun)
- Har 61-scheduling tsiklida GRQ majburiy tekshiriladi (starvation'ga qarshi)
- Work stealing — bo'sh P boshqa P'dan yarim navbatni oladi
- Go 1.14+ asinxron preemption, sysmon ~10 ms dan keyin signal yuboradi

---

### 2026-09-07 — `go.mod` va tozalash

| Commit | Nima qilindi |
|:-------|:-------------|
| `1f8fadd` | `fix: empty space removed` — ortiqcha bo'shliqlar tozalandi |
| `1472e1e` | `init: go mod` — modul `github.com/JavascriptDev347/golang-interview`, Go 1.26.2 |

---

### 2026-09-06 — Repo boshlandi

| Commit | Nima qilindi |
|:-------|:-------------|
| `1d6e239` | `init: start` — 01…05 bo'lim papkalari va `main.go` skeletlari |
| `7821343` | `first commit` — repo yaratildi |

---

## ⏭ Keyingi qadamlar / Next steps

| # | Vazifa | Bo'lim | Ustuvorlik |
|:-:|:-------|:-------|:----------:|
| 1 | `waitgroup_race` — haqiqiy race demo yozish (`-race` bilan tushishi kerak) | 02 | 🔴 yuqori |
| 2 | `waitgroup_sync` — `wg.Add/Done/Wait` to'g'ri ishlatilishi + test | 02 | 🔴 yuqori |
| 3 | `closure_bug` — loop o'zgaruvchisi bug'i (Go 1.22 gacha va keyin) | 02 | 🔴 yuqori |
| 4 | `channel/` papkasi — buffered/unbuffered, `select`, `nil` channel | 02 | 🟡 o'rta |
| 5 | `_Gcopystack` va `_Gpreempted` — qolgan 2 ta maxsus holat | 02 | 🟢 past |
| 6 | `01-fundamentals` — slice ichki tuzilishi (`len`/`cap`/`array` ko'rsatkichi) | 01 | 🟢 past |

---

## 🧩 Ochiq savollar / Open questions

| Savol | Holat |
|:------|:-----:|
| `_Gpreempted` va `_Gwaiting` orasidagi aniq farq nima? | ❓ |
| `gfree` ro'yxati qachon tozalanadi (GC bilan bog'liqmi)? | ❓ |
| Bloklovchi syscall'da M ko'payishi — amaliy limiti qancha? | ❓ |
