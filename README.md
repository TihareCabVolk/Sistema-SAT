# Sistema de Alerta Temprana de Sismos - SAT

## Descripción

Este proyecto implementa un sistema distribuido basado en una arquitectura de microservicios para la gestión de alertas tempranas de sismos.

La solución utiliza comunicación síncrona mediante REST para la recepción inicial de reportes y comunicación asíncrona basada en eventos para el procesamiento interno, permitiendo alta disponibilidad, desacoplamiento entre servicios y tolerancia a grandes volúmenes de eventos.

---

# Arquitectura del Sistema

El ecosistema se encuentra compuesto por diversos componentes, los que se pueden ver a continuación:

## Front-End



---

## Bases de Datos

Se utilizan tres bases de datos independientes, que son las siguientes:

| Servicio | Base de Datos |
|----------|---------------|
| Centro de Reportes | DB Reportes |
| Validación | DB Historial Geográfico |
| Logística | DB Costos y Alertas |

---

## Back-End

