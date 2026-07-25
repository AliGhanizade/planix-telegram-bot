# 🗂 پلنیکس | Planix Telegram Bot

بات تلگرامی برای برنامه‌ریزی روزانه و پیگیری تسک‌ها — نوشته‌شده با Go.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Telegram](https://img.shields.io/badge/Telegram-Bot-26A5E4?logo=telegram&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)

پلنیکس به هر کاربر کمک می‌کند کارهایش را ثبت کند، برنامه‌ی امروزش را ببیند و تسک‌ها را به دیگران واگذار کند و وضعیتشان را پیگیری کند. همه‌ی تعامل‌ها فارسی و با کیبورد و دکمه‌های شیشه‌ای هستند.

## ✨ قابلیت‌ها

- ➕ **ثبت تسک جدید** — فقط عنوان را بنویس؛ پلنیکس خودش موعد پیش‌فرض می‌گذارد
- 📋 **برنامه امروز** — فهرست تسک‌های باز با دکمه‌ی تیک زدن و کارت جزئیات
- 👥 **واگذاری تسک** — چند تسک را با یک پیام به یک کاربر دیگر بسپار؛ او خبردار می‌شود
- 📊 **وضعیت وظایف دیگران** — ببین هر نفر کدام تسک‌های تو را انجام داده
- ✅ **اطلاع‌رسانی دوطرفه** — وقتی مجری تسک را تیک می‌زند، مالک تسک پیام می‌گیرد
- 📝 **ویرایش و حذف** — عنوان و توضیحات هر تسک را از کارت تسک تغییر بده
- 👤 **پروفایل** — مشخصات و تعداد تسک‌های باز تو
- 🌐 **وب‌هوک تلگرام** — با اعتبارسنجی `X-Telegram-Bot-Api-Secret-Token`
- 🩺 **سلامت سرویس** — اندپوینت `/healthz` و لاگ ساخت‌یافته با zap

## 🏗 معماری

پروژه از لایه‌بندی استاندارد Go پیروی می‌کند؛ جزئیات کامل در [docs/ARCHITECTURE.fa.md](docs/ARCHITECTURE.fa.md).

```
cmd/planix/            نقطه‌ی ورود برنامه و خاموشی نرم
internal/app/          چیدن اجزا: config، logger، db، bot، scheduler
internal/domain/       فقط مدل‌های دیتابیس
internal/repository/   کوئری‌های CRUD
internal/service/      قواعد کسب‌وکار و لاگ رویدادها
internal/bot/          پیام‌ها، کیبوردها، ماشین وضعیت و کال‌بک‌های تلگرام
internal/httpapi/      روتر Gin: سلامت سرویس، وب‌هوک، OpenAPI
internal/platform/     آداپتورهای PostgreSQL و لاگر
docs/                  قرارداد OpenAPI و مستندات معماری
```

## 🚀 اجرای محلی

۱. مخزن را کلون کنید و پیش‌نیازها را نصب کنید (Go 1.25+ و Docker):

```bash
git clone https://github.com/AliGhanizade/planix-telegram-bot.git
cd planix-telegram-bot
```

۲. فایل `.env.example` را به `.env` کپی کنید و توکن واقعی بات را از [@BotFather](https://t.me/BotFather) بگذارید:

```bash
cp .env.example .env
```

۳. دیتابیس را بالا بیاورید و متغیرهای محیطی را لود کنید:

```bash
docker compose up -d postgres
export $(grep -v '^#' .env | xargs)
```

۴. اجرا:

```bash
go run ./cmd/planix
```

۵. سلامت سرویس: `http://localhost:8080/healthz`

## ⚙️ متغیرهای محیطی

| متغیر | الزامی | پیش‌فرض | توضیح |
|---|---|---|---|
| `TELEGRAM_BOT_TOKEN` | ✅ | — | توکن بات از BotFather |
| `DATABASE_URL` | ✅ | اتصال لوکال پستگرس | رشته‌ی اتصال PostgreSQL |
| `TELEGRAM_WEBHOOK_SECRET` | ❌ | خالی | برای اعتبارسنجی وب‌هوک |
| `TELEGRAM_OWNER_USERNAME` | ❌ | `AliGhanizade` | یوزرنیم پشتیبانی |
| `HTTP_ADDR` | ❌ | `:8080` | آدرس HTTP سرور |
| `APP_ENV` | ❌ | `development` | `production` یا `development` |
| `LOG_LEVEL` | ❌ | `info` | `info` یا `debug` |
| `DAILY_REPORT_CRON` | ❌ | `0 0 21 * * *` | زمان گزارش روزانه (cron با ثانیه) |

## 🐳 داکر

```bash
docker compose up -d postgres
docker build -t planix-bot .
docker run --rm --env-file .env -p 8080:8080 planix-bot
```

## 🤖 حالت وب‌هوک (پیشنهادی برای پروداکشن)

برنامه را روی HTTPS منتشر کنید و سپس:

```
https://api.telegram.org/bot<TOKEN>/setWebhook?url=https://YOUR-DOMAIN/telegram/webhook&secret_token=<TELEGRAM_WEBHOOK_SECRET>
```

اگر `TELEGRAM_WEBHOOK_SECRET` تنظیم شده باشد، هر درخواستی که هدر مخفی آن را درست نیاورَد با `401` رد می‌شود. در حالت پیش‌فرض بات با لانگ‌پولینگ کار می‌کند.

## 🏷 نسخه‌ها

- **v0.1.0** — نسخه‌ی اولیه: ثبت تسک، برنامه امروز، تیک زدن تسک
- **v0.2.0** — واگذاری تسک به دیگران، پیشنهاد کاربران، پیگیری وضعیت
- **v1.0.0** — ویرایش و حذف تسک، پروفایل و پشتیبانی، وب‌هوک، مستندات فارسی

تغییرات هر نسخه در [docs/CHANGELOG.md](docs/CHANGELOG.md) ثبت شده است.

## 🗳 نقشه‌ی راه

- [ ] گزارش روزانه‌ی خودکار (زمان‌بند آماده است)
- [ ] ویرایش موعد و اولویت از کارت تسک
- [ ] پیوست مدرک انجام تسک (عکس/فایل)
- [ ] تایم‌زون کاربر برای موعد تسک‌ها

## 📄 قرارداد API

قرارداد OpenAPI در [`docs/openapi.yaml`](docs/openapi.yaml) نگه‌داری می‌شود و در زمان اجرا روی `/openapi.yaml` در دسترس است.
