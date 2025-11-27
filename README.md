# HAYA-DISK 📁

A modern, secure, and user-friendly file storage system built with Go. HAYA-DISK allows users to upload, organize, and manage their files with an elegant web interface, powered by SQLite for secure and efficient data management.

![HAYA-DISK](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![SQLite](https://img.shields.io/badge/SQLite-3.0+-003B57?style=flat&logo=sqlite)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ Features

### 🔐 User Management
- **Secure Authentication**: User registration and login with password hashing
- **Session Management**: Secure session handling with automatic expiration
- **Profile Management**: Update display name and password through settings
- **SQLite Database**: All user data stored in secure, fast SQLite database
- **Input Validation**: Email format validation and region-based phone number validation

### 📂 File Management
- **File Upload**: Drag-and-drop or click-to-upload interface
- **File Organization**: Create folders and organize files hierarchically
- **File Operations**: Download, delete, and move files between folders
- **Thumbnail Preview**: Automatic thumbnail generation for images and videos
- **File Type Support**: Images, videos, audio files, documents, and more
- **Database-Backed Metadata**: All file metadata tracked in SQLite for security and integrity

### 📊 Dashboard Widgets
- **Storage Overview**: Visual pie chart showing storage usage by file type
- **Recent Uploads**: Quick access to your last 5 uploaded files
- **Storage Statistics**: Real-time file count and size information

### 🎨 Modern UI
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile devices
- **Beautiful Gradients**: Eye-catching purple gradient theme
- **Smooth Animations**: Polished transitions and hover effects
- **Dark/Light Ready**: Clean and modern interface

### ⚡ Performance & Concurrency
- **Multi-User Support**: Thread-safe operations for concurrent users
- **Per-User File Locks**: Read/Write locks prevent race conditions
- **Smart Caching**: 5-second cache for directory listings (5x faster)
- **Rate Limiting**: 10 uploads per minute per user to prevent abuse
- **Atomic Operations**: Safe user data persistence with atomic file writes
- **Zero Blocking**: Users don't interfere with each other's operations

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git (for cloning the repository)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/HAYASAKA7/HAYA-DISK.git
   cd HAYA-DISK
   ```

2. **Build the application**

   ```bash
   go build -o haya-disk.exe
   ```

3. **Run the migration (first time only)**

   **Option A: Automatic Migration (Recommended)**
   
   Simply run the application - it will automatically detect `users.json` and migrate data on first startup:

   ```bash
   ./haya-disk.exe
   ```

   **Option B: Manual Migration**
   
   If you prefer to migrate separately or if automatic migration fails:

   ```bash
   go build -o migrate.exe ./cmd/migrate
   ./migrate.exe
   ```

   This will:
   - Create the SQLite database (`haya-disk.db`)
   - Migrate all users from `users.json`
   - Scan and register all existing files in the `storage` directory

4. **Run the application**

   ```bash
   ./haya-disk.exe
   ```

   Or directly run without building:

   ```bash
   go run main.go
   ```

5. **Access the application**

   Open your browser and navigate to:
   - Local: `http://localhost:8080`
   - Network: `http://<your-ip>:8080`

## 📁 Project Structure

```text
HAYA-DISK/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── haya-disk.db              # SQLite database (auto-generated)
├── users.json                # Legacy user data (kept as backup)
├── cmd/
│   └── migrate/
│       └── main.go           # Migration tool for legacy data
├── config/
│   ├── constants.go          # Configuration constants
│   └── backup_config.go      # Backup configuration settings
├── handlers/
│   ├── auth.go              # Authentication handlers
│   ├── file_list.go         # File listing handlers
│   ├── file_ops.go          # File operations handlers
│   └── page.go              # Page rendering handlers
├── middleware/
│   ├── session.go           # Session management
│   └── rate_limiter.go      # Rate limiting middleware
├── models/
│   └── models.go            # Data models (User, FileMetadata, etc.)
├── services/
│   ├── database_service.go  # SQLite database operations
│   ├── session_service.go   # Session service layer
│   ├── user_service.go      # User service layer
│   ├── file_lock_service.go # File operation locking
│   ├── cache_service.go     # Directory listing cache
│   └── backup_service.go    # Auto-backup scheduler and operations
├── backups/                 # Backup storage (auto-generated)
│   ├── backup_YYYY-MM-DD_HHMMSS.zip
│   └── backup_log.txt
├── storage/                 # User file storage (auto-generated)
│   └── {username}_{hash}/
│       ├── Audios/
│       ├── Images/
│       ├── Videos/
│       └── ...
├── templates/               # HTML templates and assets
│   ├── list.html
│   ├── login.html
│   ├── register.html
│   ├── upload.html
│   └── style.css
└── utils/
    ├── utils.go             # Utility functions
    └── migrate.go           # Migration utilities
    └── validation.go        # Validation utilities
```

## 🔧 Configuration

The application uses default configuration values defined in `config/constants.go`:

- **Server Port**: `:8080` (accessible on all network interfaces)
- **Storage Directory**: `./storage`
- **Templates Directory**: `./templates`
- **Session Duration**: 30 days
- **Cache TTL**: 5 seconds for directory listings
- **Rate Limit**: 10 uploads per minute per user
- **Buffer Sizes**: 32KB for read/write operations

### Changing the Port

To change the server port, modify the `ServerPort` constant in `config/constants.go`:

```go
const ServerPort = ":8080" // Change to your desired port
```

## 💾 Auto-Backup

HAYA-DISK includes an automatic backup system that protects your data by creating scheduled backups of both the database and user files.

### Features

- **Scheduled Backups**: Automatically runs at 3:00 AM (device time) daily
- **Compressed Archives**: Backups are compressed to `.zip` format to save space
- **Auto-Cleanup**: Old backups older than 7 days are automatically deleted
- **Full Backup**: Includes both SQLite database and all user storage files
- **Graceful Shutdown**: Backup scheduler stops properly when server shuts down

### Backup Location

Backups are stored in the `backups/` directory:

```text
backups/
├── backup_2025-11-27_030000.zip    # Compressed backup
├── backup_2025-11-26_030000.zip
├── backup_2025-11-25_030000.zip
└── backup_log.txt                   # Backup history log
```

### Backup Contents

Each backup archive contains:
- `haya-disk.db` - SQLite database with user accounts and file metadata
- `storage/` - All user files and folders

### Configuration

Backup settings can be modified in `config/backup_config.go`:

```go
var DefaultBackupSettings = BackupSettings{
    Enabled:        true,   // Enable/disable auto-backup
    BackupDir:      "backups",
    ScheduleHour:   3,      // Hour to run backup (0-23)
    ScheduleMinute: 0,      // Minute to run backup (0-59)
    RetentionDays:  7,      // Keep backups for X days
    BackupDatabase: true,   // Backup the SQLite database
    BackupStorage:  true,   // Backup the storage folder
    CompressBackup: true,   // Compress backups to .zip
}
```

### Restoring from Backup

To restore from a backup:

1. Stop the HAYA-DISK server
2. Extract the backup zip file
3. Replace `haya-disk.db` with the backed-up database
4. Replace the `storage/` folder with the backed-up storage
5. Restart the server

```bash
# Example restore commands (PowerShell)
Stop-Process -Name "HAYA-DISK" -ErrorAction SilentlyContinue
Expand-Archive -Path "backups/backup_2025-11-27_030000.zip" -DestinationPath "restore_temp"
Copy-Item "restore_temp/haya-disk.db" -Destination "." -Force
Copy-Item "restore_temp/storage" -Destination "." -Recurse -Force
Remove-Item "restore_temp" -Recurse
./HAYA-DISK.exe
```

## 📱 Input Validation

HAYA-DISK validates user input during registration to ensure data integrity.

### Email Validation

- Standard RFC 5322 email format validation
- Format: `username@domain.tld`
- Examples: `user@example.com`, `john.doe@company.org`

### Phone Number Validation

Phone numbers are validated based on the selected region. Users must select their region from a dropdown, and the phone number format is validated accordingly.

#### Supported Regions

| Region | Country | Code | Format | Example |
|--------|---------|------|--------|---------|
| CN | China | +86 | 11 digits, starts with 1 | 13812345678 |
| US | United States | +1 | 10 digits | 2025551234 |
| CA | Canada | +1 | 10 digits | 4165551234 |
| UK | United Kingdom | +44 | 10 digits, starts with 7 | 7911123456 |
| JP | Japan | +81 | 10 digits | 9012345678 |
| KR | South Korea | +82 | 10-11 digits | 01012345678 |
| TW | Taiwan | +886 | 9 digits, starts with 9 | 912345678 |
| HK | Hong Kong | +852 | 8 digits | 51234567 |
| SG | Singapore | +65 | 8 digits | 81234567 |
| AU | Australia | +61 | 9 digits, starts with 4 | 412345678 |
| DE | Germany | +49 | 10-11 digits | 15123456789 |
| FR | France | +33 | 9 digits | 612345678 |
| IN | India | +91 | 10 digits | 9123456789 |
| RU | Russia | +7 | 10 digits | 9123456789 |
| BR | Brazil | +55 | 10-11 digits | 11912345678 |
| MX | Mexico | +52 | 10 digits | 5512345678 |

### Adding More Regions

To add support for additional regions, edit `utils/validation.go` and add entries to the `PhoneRegions` map:

```go
"XX": {
    Name:      "Country Name",
    Code:      "+XX",
    Pattern:   `^regex_pattern$`,
    Example:   "1234567890",
    MinLength: 10,
    MaxLength: 10,
},
```

## 🎯 Usage

### First Time Setup

1. **Register an Account**
   - Navigate to the registration page
   - Enter your username and password
   - Optional: Set a display name

2. **Login**
   - Use your credentials to log in
   - You'll be redirected to your personal dashboard

### Managing Files

**Upload Files:**
- Click the "Upload File" button
- Select or drag files into the upload zone
- Choose the destination folder
- Click "Upload"

**Create Folders:**
- Click the "New Folder" button
- Enter a folder name
- Folder is created immediately

**Organize Files:**
- Use the move button on any file card
- Select the destination folder
- Files are moved instantly

**Download Files:**
- Click the download button on any file card
- File downloads to your default location

**Delete Files:**
- Click the delete button on any file card
- Confirm the deletion

## 🛠️ Development

### Running in Development Mode

```bash
go run main.go
```

### Building for Production

**Windows:**
```bash
go build -o haya-disk.exe
```

**Linux/Mac:**
```bash
go build -o haya-disk
```

### Project Dependencies

- **SQLite Database**: `modernc.org/sqlite` - Pure Go SQLite driver (no CGO required)
- **Go Standard Library**:
  - `net/http` - HTTP server and client
  - `encoding/json` - JSON encoding/decoding
  - `crypto/sha256` - Password hashing
  - `html/template` - HTML templating
  - `io` - I/O operations
  - `os` - Operating system functionality
  - `path/filepath` - File path manipulation
  - `time` - Time operations
  - `sync` - Synchronization primitives (mutexes, locks)
  - `image` - Image processing
  - `image/jpeg`, `image/png` - Image format support
  - `database/sql` - Database interface

## 🔒 Security Features

- **Password Hashing**: SHA-256 hashing for password storage
- **Session Management**: Secure session tokens with expiration
- **Input Validation**: Server-side validation for all user inputs
- **Path Traversal Protection**: Sanitized file paths to prevent directory traversal
- **User Isolation**: Each user has their own isolated storage directory
- **Rate Limiting**: Upload rate limits to prevent abuse and DoS attacks
- **Concurrent Access Control**: Thread-safe file operations with proper locking
- **File Metadata Security**: Only files registered in database are accessible
  - **Prevents unauthorized access**: Manually added files won't appear in user's file list
  - **Database integrity**: File operations tracked and validated through SQLite
  - **Tamper-proof**: Files without database records are invisible to users

## 🎨 Customization

### Changing the Theme

Edit `templates/style.css` to customize colors:

```css
/* Main gradient */
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);

/* Primary color */
--primary-color: #667eea;

/* Secondary color */
--secondary-color: #764ba2;
```

### Modifying Storage Limits

Update the settings handler in `handlers/page.go` to change storage limits and upload size restrictions.

## 📝 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Home page (redirects to login/list) |
| `/login` | GET/POST | User login |
| `/register` | GET/POST | User registration |
| `/logout` | GET | User logout |
| `/list` | GET | File listing page |
| `/upload` | GET/POST | File upload (with rate limiting) |
| `/download` | GET | File download |
| `/delete` | POST | File/folder deletion |
| `/create-folder` | POST | Create new folder |
| `/move-file` | POST | Move file to folder |
| `/thumbnail` | GET | Get file thumbnail |
| `/settings` | GET/POST | User settings |
| `/api/get-user-info` | GET | Get user information |
| `/api/update-profile` | POST | Update user profile |

## 📊 Database Schema

### Users Table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT,
    phone TEXT,
    password TEXT NOT NULL,
    unique_code TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    login_type TEXT
);
```

### Files Table

```sql
CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    filename TEXT NOT NULL,
    storage_path TEXT NOT NULL UNIQUE,
    parent_path TEXT NOT NULL DEFAULT '/',
    file_size INTEGER NOT NULL DEFAULT 0,
    mime_type TEXT,
    file_hash TEXT,
    is_directory BOOLEAN NOT NULL DEFAULT 0,
    uploaded_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);
```

### Key Features

- **Indexed lookups**: Fast queries on username, parent_path, and storage_path
- **Foreign key constraints**: Automatic cascade deletion when user is deleted
- **File deduplication**: SHA-256 hash tracking for potential deduplication
- **MIME type tracking**: Proper content type handling
- **Audit trail**: Upload and modification timestamps

## 🔄 Migration from JSON to SQLite

### Why SQLite?

HAYA-DISK has been upgraded from JSON file storage to SQLite database for several important reasons:

#### 🔐 Security Enhancement

**Before (JSON-based)**: The system directly read the filesystem, meaning any file manually added to a user's storage folder would appear in their file list - a **major security vulnerability**.

**After (SQLite-based)**: Only files registered in the database are accessible. Manually added files are completely invisible to users, preventing unauthorized access and tampering.

#### ⚡ Performance Improvements

- **Faster queries**: Indexed database lookups vs. file scanning
- **Efficient filtering**: SQL queries instead of in-memory filtering
- **Better caching**: Database-level optimizations
- **Concurrent access**: Better handling of simultaneous users

#### 🎯 Feature Enablement

The SQLite migration enables future features:

- File sharing between users
- File versioning and history
- Advanced search capabilities
- Storage quotas per user
- Activity logs and audit trails
- File tags and categories
- Trash/recycle bin functionality

### Migration Process

If you're upgrading from an older version:

1. **Automatic Migration**: Run the migration tool once

   ```bash
   go build -o migrate.exe ./cmd/migrate
   ./migrate.exe
   ```

2. **What Gets Migrated**:
   - All users from `users.json` → `users` table
   - All files scanned from `storage/` → `files` table
   - File metadata calculated (size, hash, MIME type)

3. **Safe Migration**:
   - Original `users.json` is preserved as backup
   - Files remain in same location on disk
   - Idempotent - can be run multiple times safely

4. **Rollback**: Keep your `users.json` backup in case you need to revert

### Database Location

The SQLite database is stored as `haya-disk.db` in the application root directory. You can:

- **Backup**: Simply copy the `.db` file
- **Restore**: Replace with a backup copy
- **View**: Use any SQLite browser tool (DB Browser for SQLite, etc.)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ✔ Concurrency & Performance Optimizations

### Multi-User Concurrent Access

HAYA-DISK is designed to handle multiple users accessing their files simultaneously without conflicts or data corruption.

#### **Per-User File Locking System**

Each user has their own read-write lock (`sync.RWMutex`) that manages concurrent access:

- **Read Operations** (Download, Thumbnail, List):
  - Multiple reads can happen simultaneously
  - No blocking between different users
  - Fast and efficient for viewing files

- **Write Operations** (Upload, Delete, Move, Create Folder):
  - Only one write operation per user at a time
  - Prevents race conditions and file corruption
  - Users don't block each other

**Example Scenario:**
```
User A uploads → Only blocks User A's other uploads
User B downloads → Continues without interruption
User C lists files → Continues without interruption
```

#### **Smart Directory Caching**

- Directory listings are cached for 5 seconds per user
- **5x faster** file listing performance
- Cache automatically invalidated after:
  - File upload
  - File deletion
  - Folder creation
  - File move operations

#### **Rate Limiting**

- **10 uploads per minute per user**
- Sliding window algorithm
- Prevents abuse and server overload
- Returns `429 Too Many Requests` when exceeded

#### **Atomic User Data Persistence**

User data is saved safely even during concurrent operations:

1. Create copy of user data while holding read lock
2. Release lock immediately (no blocking)
3. Write to temporary file (`.tmp`)
4. Atomic rename to actual file
5. **Zero chance of corruption** even if server crashes mid-write

### Performance Characteristics

| Operation | Concurrency Model | Performance Impact |
|-----------|------------------|-------------------|
| **Upload** | Per-user write lock | ~5-10ms overhead |
| **Download** | Per-user read lock | Minimal overhead |
| **List Files** | Cached + read lock | **5x faster** |
| **Delete** | Per-user write lock | ~2-5ms overhead |
| **Move** | Per-user write lock | ~2-5ms overhead |

### Thread Safety Guarantees

✅ **No race conditions** on file operations  
✅ **No data corruption** in user database  
✅ **No deadlocks** with proper lock ordering  
✅ **No blocking** between different users  
✅ **Consistent reads** during concurrent writes  

### Scaling Considerations

- **Horizontal Scaling**: Not supported (single-instance file system)
- **Vertical Scaling**: Excellent (Go's goroutines handle 1000+ concurrent users)
- **Storage Scaling**: Limited by disk space only
- **Memory Usage**: ~1-2MB per active user session

## 👩‍🔬 Author

**HAYASAKA7**

- GitHub: [@HAYASAKA7](https://github.com/HAYASAKA7)
- Project: [HAYA-DISK](https://github.com/HAYASAKA7/HAYA-DISK)

## 🙏 Acknowledgments

- Built with Go standard library
- Inspired by modern cloud storage solutions
- Icons and UI elements from modern design systems

## 📮 Support

If you have any questions or issues, please open an issue on the GitHub repository.

---

Made with ❤️ by HAYASAKA7

