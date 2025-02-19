# TarantulaGo 🕷️

A Telegram bot for managing tarantula collections and cricket colonies. Track molts, feedings, health status, and maintain breeding colonies all through an easy-to-use Telegram interface.

## Features

- 🕷️ **Tarantula Management**
  - Track individual tarantulas with detailed profiles
  - Record molts and monitor growth
  - Schedule and track feedings
  - Monitor health status
  - Set up custom feeding schedules based on species and size

- 🦗 **Cricket Colony Management**
  - Track multiple cricket colonies
  - Monitor colony population
  - Record feeding usage
  - Get low colony alerts
  - Track colony sustainability

- 🔔 **Notifications**
  - Customizable feeding reminders
  - Colony status alerts
  - Health check reminders
  - Molt monitoring alerts

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Telegram Bot Token

## Environment Variables

```env
TELEGRAM_BOT_TOKEN=your_bot_token
POSTGRES_URL=postgresql://username:password@localhost:5432/dbname?sslmode=disable
LOG_LEVEL=info  # Optional, defaults to "info"
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/tarantulago.git
cd tarantulago
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
- Create a PostgreSQL database
- The application will automatically create the required schema and tables on first run

4. Run the application:
```bash
go run main.go
```

## Database Schema

The application uses a PostgreSQL database with the `spider_bot` schema. Tables include:
- `tarantulas` - Main tarantula information
- `molt_records` - Molt history
- `feeding_events` - Feeding records
- `cricket_colonies` - Cricket colony management
- `health_check_records` - Health monitoring
- And more supporting tables

## Usage

1. Start a chat with your bot on Telegram
2. Use the `/start` command to initialize
3. Navigate through the menu to:
   - Add new tarantulas
   - Record feedings
   - Track molts
   - Manage cricket colonies
   - Configure notifications

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [telebot](https://github.com/tucnak/telebot)
- Uses [GORM](https://gorm.io) for database operations

## Support

For support, please open an issue in the GitHub repository or contact the maintainers.
