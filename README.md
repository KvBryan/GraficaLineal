# 📊 GraficaLineal - Visualización Científica con DeepSeek & Go

Aplicación web ligera de alto rendimiento diseñada específicamente para la generación, auditoría y exportación de gráficos científicos para **documentos de investigación, tesis académicas y artículos de revistas (papers)**.

Construida con **Go (Golang)** bajo **Vertical Slice Architecture (VSA)**, integrada con el modelo de lenguaje de última generación **DeepSeek V3 / Chat**, y visualización interactiva basada en **Apache ECharts 5**.

---

## 🚀 Características Principales

### 1. Soporte para Documentos Excel, Word y PDF
- **Documentos de Microsoft Word (`.docx`)**: Parser nativo en Go que inspecciona el documento, detecta tablas y bloques de texto cuantitativos, y permite al usuario elegir qué tabla o sección graficar mediante preguntas interactivas.
- **Libros de Microsoft Excel (`.xlsx`, `.xls`)**: Análisis automático de hojas de cálculo, columnas numéricas y dimensiones, con previsualización inmediata y edición en vivo.
- **Artículos y Manuscritos en PDF (`.pdf`)**: Extracción de páginas, detección de tablas cuantitativas alineadas y selección interactiva de secciones para análisis con DeepSeek.

### 2. Catálogo Completo de 13 Tipos de Gráficas Científicas

| Categoría | Tipo de Gráfica | Descripción y Caso de Uso |
| :--- | :--- | :--- |
| **1. Comparación entre categorías** | **Barras / Columnas** | Comparación de medias, conteos y resultados discretos por grupo. |
| | **Radial / Araña (Radar)** | Evaluación multidimensional de variables simultáneas en la misma escala. |
| **2. Evolución y cambio en el tiempo** | **Líneas** | Tendencias y evolución temporal de variables continuas. |
| | **Áreas** | Magnitud acumulada y volumen a lo largo del tiempo. |
| | **Diagrama de Gantt** | Cronogramas, hitos y fases metodológicas de proyectos. |
| **3. Composición y partes de un todo** | **Circular (Pastel)** | Proporciones porcentuales de un total (3 a 6 categorías). |
| | **Anillos (Dona)** | Variante circular con espacio central para métricas destacadas. |
| | **Embudo (Funnel)** | Etapas de tamizaje, selección de muestra o conversión. |
| | **Cascada (Waterfall)** | Descomposición de variaciones incrementales positivas y negativas. |
| **4. Distribución y relación** | **Dispersión (Scatter plot)** | Correlación entre dos variables cuantitativas con puntos individuales. |
| | **Histograma** | Frecuencias agrupadas en intervalos de clase continuos (bins). |
| | **Diagrama de Caja (Box plot)** | Resumen Tukey: Mediana, cuartiles (Q1, Q3), rango intercuartílico y atípicos. |
| | **Mapa de Calor (Heatmap)** | Matrices de correlación o densidad bidimensional de datos. |

### 3. Trazabilidad Académica Total ("Qué tablas usó y cómo")
- **Auditoría de Fuentes**: Muestra con exactitud qué hoja, rango de celdas o tabla de Word alimentó la figura.
- **Metodología y Razonamiento DeepSeek**: Explicación formal de los cálculos matemáticos aplicados (sumas, porcentajes, medianas).
- **Editor en Tiempo Real**: Hoja de cálculo editable incrustada en la interfaz; cualquier cambio numérico actualiza la gráfica al instante.
- **Exportación en Excel (`.xlsx`)**: Genera un archivo con dos hojas:
  - Hoja 1: `Datos de la Gráfica` (datos tabulados limpios con formato formal).
  - Hoja 2: `Metodología y Fuentes` (registro metodológico detallado y nota de cita para el artículo).
- **Exportación de Imagen en Alta Resolución**: Descarga directa de la figura en PNG (High DPI) lista para insertar en el manuscrito.

---

## 🏛️ Arquitectura: Vertical Slice Architecture (VSA)

El backend en Go implementa una arquitectura por rebanadas verticales (Vertical Slices), eliminando el acoplamiento excesivo y organizando el código por capacidades de negocio:

```
GraficaLineal/
├── cmd/server/main.go               # Punto de entrada HTTP y arranque
├── internal/
│   ├── config/config.go             # Carga de variables de entorno
│   ├── core/domain/                 # Entidades y tipos compartidos (ChartSpec, TableData)
│   └── slices/
│       ├── document_analysis/       # Slice 1: Inspección de Excel y Word
│       ├── ai_generation/           # Slice 2: Integración con DeepSeek API
│       ├── chart_editor/            # Slice 3: Transformación y cálculos en tiempo real
│       └── excel_export/            # Slice 4: Generación de libro Excel de investigación
├── web/
│   ├── static/                      # CSS científico, JS reactivo y ECharts 5 bundle
│   └── index.html                   # Interfaz de usuario SPA
├── Dockerfile                       # Construcción multi-stage ultraligera
├── docker-compose.yml               # Orquestación de contenedores
└── .env.example                     # Plantilla de configuración
```

---

## 🛠️ Ejecución Local con Docker

### Prerrequisitos
- Docker y Docker Compose instalados.

### 1. Iniciar el Contenedor
```bash
docker compose up -d --build
```

### 2. Acceder a la Aplicación
Abre tu navegador en:
```
http://localhost:8090
```

### 3. Detener la Aplicación
```bash
docker compose down
```

---

## 💻 Ejecución Local con Go (Sin Docker)

```bash
# Descargar dependencias
go mod download

# Iniciar el servidor
go run cmd/server/main.go
```

---

## ☁️ Despliegue en la Nube (Producción / GitHub)

La aplicación está completamente dockerizada y lista para desplegarse en plataformas como:
- **Render** / **Railway** / **Fly.io**: Conectar el repositorio de GitHub y seleccionar Docker.
- **VPS / Servidor Linux**: Clonar el repositorio y ejecutar `docker compose up -d`.

### Variables de Entorno
| Variable | Descripción | Valor por Defecto |
| :--- | :--- | :--- |
| `PORT` | Puerto HTTP del servidor | `8090` |
| `DEEPSEEK_API_KEY` | Llave API oficial de DeepSeek | Configurada en `.env` |
| `DEEPSEEK_BASE_URL` | URL base de la API | `https://api.deepseek.com` |
| `DEEPSEEK_MODEL` | Modelo de lenguaje | `deepseek-chat` |

---

## 📄 Licencia
Este proyecto está bajo la Licencia MIT. Consulta el archivo `LICENSE` para más detalles.
