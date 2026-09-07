# 🟠 `_Gsyscall` — system call ichida turgan goroutine

<div align="center">

⬅️ [Ortga: `_Grunning`](./Grunning.md) · 🏠 [Repo bosh sahifa](../../README.md) · ➡️ [Keyingi: `_Gwaiting`](./Gwaiting.md)

</div>

> **EN** · *"This goroutine asked the operating system to do something (read a file, call C code) and is now waiting inside the kernel. Its thread is stuck — but Go makes sure the rest of the program keeps running."*
> **UZ** · *"Bu goroutine operatsion tizimdan biror ish so'radi (fayl o'qish, C kodini chaqirish) va hozir kernel ichida kutmoqda. Uning thread'i qotib qoldi — lekin Go dasturning qolgan qismi ishlashda davom etishini ta'minlaydi."*

---

## 1️⃣ `_Gsyscall` nima? / What is it?

| 🇬🇧 English | 🇺🇿 O'zbekcha |
|:-----------|:-------------|
| The goroutine called into the operating system and control left Go's world. | Goroutine operatsion tizimga murojaat qildi va boshqaruv Go dunyosidan chiqib ketdi. |
| The **M (OS thread) is blocked** in the kernel — it cannot run any other goroutine. | **M (OS thread) kernel ichida bloklangan** — u boshqa hech qanday goroutine'ni bajara olmaydi. |
| The G stays **attached to that M**, but the **P is given away** so other goroutines keep running. | G o'sha **M'ga bog'langan** qoladi, lekin **P boshqaga beriladi** — shunda qolgan goroutine'lar ishlashda davom etadi. |
| This is the key trick: one blocked thread does **not** block the whole program. | Asosiy hiyla shu: bitta bloklangan thread butun dasturni **bloklamaydi**. |
| Because of this, the number of Ms can grow **far beyond `GOMAXPROCS`** (limit: 10 000 threads). | Shu sababli M'lar soni **`GOMAXPROCS`dan ancha ko'p** bo'lishi mumkin (chegara: 10 000 thread). |

### 🧠 Analogiya / Analogy

> **EN** · The cashier goes down to the warehouse with a customer to fetch a rare item. The cashier (M) and customer (G) are both gone for a while — so the shop hands the **cash register (P)** to another cashier, and the queue keeps moving.
>
> **UZ** · Kassir mijoz bilan birga noyob mahsulot olib kelish uchun omborga tushib ketdi. Kassir (M) ham, mijoz (G) ham bir muddat yo'q — shuning uchun do'kon **kassani (P)** boshqa kassirga beradi va navbat to'xtamaydi.

---

## 2️⃣ 🗺 Model chizmasi / Model

```text
   1️⃣ SYSCALL'GA KIRISH — entersyscall()

   ┌──────────────────────────┐
   │  M1  +  P0  +  G5        │   G5 → os.ReadFile(...) chaqirdi
   │  🟢 _Grunning            │
   └────────────┬─────────────┘
                ▼
   ┌──────────────────────────┐        ┌────────────────────────┐
   │  M1  +  G5               │        │   P0  bo'shatildi      │
   │  🟠 _Gsyscall            │  ────▶ │   (_Psyscall holati)   │
   │  kernel ichida qotgan    │        │   LRQ: G1 G2 G3 …      │
   └──────────────────────────┘        └────────────────────────┘


   2️⃣ AGAR SYSCALL UZOQ CHO'ZILSA (>20 µs) — sysmon aralashadi

   ┌──────────────────────────┐        ┌────────────────────────┐
   │  M1  +  G5               │        │   P0  →  M2 ga berildi │
   │  🟠 hali ham kernel'da   │        │   👮 sysmon retake()   │
   └──────────────────────────┘        └───────────┬────────────┘
                                                   ▼
                                       ┌────────────────────────┐
                                       │  M2 + P0 + G1          │
                                       │  🟢 ish davom etmoqda  │
                                       └────────────────────────┘

   3️⃣ SYSCALL QAYTGANDA — exitsyscall()

   ┌──────────────────────────┐
   │   M1: "menga P kerak!"   │
   └────────────┬─────────────┘
                │
        ┌───────┴────────┐
        │                │
   ✅ bo'sh P bor    ❌ bo'sh P yo'q
        │                │
        ▼                ▼
   ┌──────────┐   ┌──────────────────────────┐
   │_Grunning │   │ G5 → _Grunnable → GRQ    │
   │ davom    │   │ M1 → park (uxlaydi)      │
   └──────────┘   └──────────────────────────┘
```

```mermaid
stateDiagram-v2
    direction LR
    _Grunning --> _Gsyscall: entersyscall()
    _Gsyscall --> _Grunning: exitsyscall() — P topildi
    _Gsyscall --> _Grunnable: exitsyscall() — P yo'q, GRQ'ga
```

---

## 3️⃣ Tez va sekin syscall / Fast vs slow syscall

| | ⚡ Tez syscall | 🐢 Sekin syscall |
|:--|:--------------|:-----------------|
| Davomiyligi | < 20 µs | > 20 µs |
| P'ga nima bo'ladi | `_Psyscall` holatida **band turadi** — M uni qaytarib oladi | 👮 `sysmon` P'ni **tortib oladi** (`retake`) va boshqa M'ga beradi |
| Yangi thread | Kerak emas | Kerak bo'lishi mumkin (yangi M yaratiladi) |
| Narxi | Deyarli tekin | Qimmatroq — thread almashinuvi bor |
| Misol | `time.Now()` ba'zi tizimlarda | Katta fayl o'qish, `cgo` chaqiruvi |

> **EN** · Go optimistically assumes syscalls are fast and only pays the handoff cost when `sysmon` notices the call is dragging on.
> **UZ** · Go optimistik tarzda syscall'lar tez deb hisoblaydi va faqat `sysmon` cho'zilib ketganini sezgandagina P'ni uzatish xarajatini to'laydi.

---

## 4️⃣ ⚠️ Eng muhim farq: tarmoq I/O `_Gsyscall` EMAS

Bu — intervyuda eng ko'p adashiladigan joy.

| Operatsiya / Operation | Holat / State | Nima uchun / Why |
|:-----------------------|:-------------:|:-----------------|
| `os.ReadFile()`, fayl o'qish | 🟠 `_Gsyscall` | Fayl I/O'ni asinxron qilib bo'lmaydi → thread bloklanadi |
| `cgo` orqali C funksiyasi | 🟠 `_Gsyscall` | C kodi Go scheduler'ini bilmaydi |
| `syscall.Exec`, `os/exec` | 🟠 `_Gsyscall` | To'g'ridan-to'g'ri kernel chaqiruvi |
| `conn.Read()`, HTTP so'rov | 🔵 [`_Gwaiting`](./Gwaiting.md) | **netpoller** (epoll/kqueue/IOCP) ishlatiladi — thread bloklanmaydi! |
| `<-ch`, `mu.Lock()`, `Sleep` | 🔵 [`_Gwaiting`](./Gwaiting.md) | Sof Go bloklanishi, kernel umuman aralashmaydi |

### 🌐 Netpoller nima qiladi?

| 🇬🇧 English | 🇺🇿 O'zbekcha |
|:-----------|:-------------|
| Go puts every network socket into **non-blocking** mode. | Go har bir tarmoq soketini **bloklamaydigan** rejimga o'tkazadi. |
| When data isn't ready, the goroutine parks as `_Gwaiting` instead of blocking a thread. | Ma'lumot tayyor bo'lmasa, goroutine thread'ni bloklamasdan `_Gwaiting` bo'lib to'xtaydi. |
| The netpoller (epoll on Linux) watches all sockets at once and wakes the right goroutine. | Netpoller (Linux'da epoll) barcha soketlarni bir vaqtda kuzatadi va kerakli goroutine'ni uyg'otadi. |
| **This is why Go can handle 100 000 connections with just a few threads.** | **Aynan shuning uchun Go bir necha thread bilan 100 000 ta ulanishni eplay oladi.** |

---

## 5️⃣ Intervyu savollari / Interview questions

| # | Savol / Question | Kalit javob / Key answer |
|:-:|:-----------------|:-------------------------|
| 1 | Bloklovchi syscall butun dasturni to'xtatib qo'yadimi? | Yo'q — P bo'shatilib boshqa M'ga beriladi, qolgan goroutine'lar ishlashda davom etadi. |
| 2 | M'lar soni `GOMAXPROCS`dan ko'p bo'la oladimi? | Ha — aynan `_Gsyscall` tufayli. Chegara `runtime/debug.SetMaxThreads`, standarti **10 000**. |
| 3 | HTTP so'rovi kutayotgan goroutine qaysi holatda? | 🔵 `_Gwaiting`, `_Gsyscall` emas — netpoller ishlatiladi. |
| 4 | `sysmon` qachon P'ni tortib oladi? | Syscall **20 µs**dan uzoq cho'zilsa (`retake` funksiyasi). |
| 5 | Syscall'dan qaytganda P topilmasa nima bo'ladi? | G `_Grunnable` bo'lib GRQ'ga tushadi, M esa park qilinadi (uxlaydi). |

---

<div align="center">

⬅️ [Ortga: `_Grunning`](./Grunning.md) · 🏠 [Repo bosh sahifa](../../README.md) · ➡️ [Keyingi: `_Gwaiting`](./Gwaiting.md)

</div>
