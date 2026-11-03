# Remote Pouch from Couch

**Remote Pouch from Couch** is a Google Drive-like cloud file storage application.

## About

---

The application is built as a multi-course project for Faculty of Technical Sciences, University of Novi Sad.

It's built by students:

- Lazar Bošnjak
- Kristijan Đeri
- Vladimir Čupić

The project is designed to teach students the fundamentals of:

- Service oriented architecture
- NoSQL databases
- Information security
- Methodology of software development

## Installation

---

**Requirements:**

- Docker Compose
- NodeJS v24+ and npm

```bash
git clone https://github.com/lazarbosnjak/cloud-storage-microservice.git

# Backend services
cd cloud-storage-microservice/src/backend
docker compose up -d

# Frontend application
cd cloud-storage-microservice/src/frontend
npm start
```
