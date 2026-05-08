package esclient

// Маппинги индексов: profiles и vacancies.
// Поля skill_slugs хранятся как keyword[] для exact-match по AND-семантике.
// Текстовые поля разбираются русским анализатором — фамилия «Иванов» матчит «иванова».

const ProfilesMapping = `{
  "settings": {
    "analysis": {
      "analyzer": {
        "ru_text": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "russian_stop", "russian_stemmer"]
        }
      },
      "filter": {
        "russian_stop": {"type": "stop", "stopwords": "_russian_"},
        "russian_stemmer": {"type": "stemmer", "language": "russian"}
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "first_name": {"type": "text", "analyzer": "ru_text"},
      "last_name": {"type": "text", "analyzer": "ru_text"},
      "profession_category": {"type": "text", "analyzer": "ru_text", "fields": {"keyword": {"type": "keyword"}}},
      "education_institution": {"type": "text", "analyzer": "ru_text"},
      "description": {"type": "text", "analyzer": "ru_text"},
      "role": {"type": "keyword"},
      "skill_slugs": {"type": "keyword"},
      "age": {"type": "integer"},
      "email": {"type": "keyword"},
      "tg": {"type": "keyword"},
      "avatar_id": {"type": "keyword"},
      "resume_id": {"type": "keyword"}
    }
  }
}`

const VacanciesMapping = `{
  "settings": {
    "analysis": {
      "analyzer": {
        "ru_text": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "russian_stop", "russian_stemmer"]
        }
      },
      "filter": {
        "russian_stop": {"type": "stop", "stopwords": "_russian_"},
        "russian_stemmer": {"type": "stemmer", "language": "russian"}
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "title": {"type": "text", "analyzer": "ru_text"},
      "position_status": {"type": "keyword"},
      "schedule": {"type": "keyword"},
      "work_format": {"type": "keyword"},
      "company_id": {"type": "keyword"},
      "skill_slugs": {"type": "keyword"},
      "experience": {"type": "integer"},
      "salary": {"type": "integer"},
      "create_at": {"type": "keyword"},
      "attachment_id": {"type": "keyword"}
    }
  }
}`
