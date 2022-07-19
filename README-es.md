# Weave Scope - Solución de problemas y monitoreo para Docker y Kubernetes

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master)
[![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope)
[![Go Report Card](https://goreportcard.com/badge/github.com/weaveworks/scope)](https://goreportcard.com/report/github.com/weaveworks/scope)
[![Docker Pulls](https://img.shields.io/docker/pulls/weaveworks/scope.svg?maxAge=604800)](https://hub.docker.com/r/weaveworks/scope/)
[![GoDoc](https://godoc.org/github.com/weaveworks/scope?status.svg)](https://godoc.org/github.com/weaveworks/scope)
[![Good first issues](https://img.shields.io/github/issues/weaveworks/scope/good-first-issue.svg?color=blueviolet\&label=good%20first%20issues)](https://github.com/weaveworks/scope/issues?q=is%3Aissue+is%3Aopen+label%3Agood-first-issue)

Weave Scope genera automáticamente un mapa de su aplicación, lo que le permite
comprenda, supervise y controle de forma intuitiva su aplicación basada en microservicios en contenedores.

## Comprenda sus contenedores Docker en tiempo real

<img src="imgs/topology.png" width="200" alt="Map you architecture" align="right">

Elija una descripción general de su infraestructura de contenedores o concéntrese en un microservicio específico. Identifique y corrija fácilmente los problemas para garantizar la estabilidad y el rendimiento de sus aplicaciones en contenedores.

## Detalles contextuales y enlaces profundos

<img src="imgs/selected.png" width="200" alt="Focus on a single container" align="right">

Vea métricas contextuales, etiquetas y metadatos para sus contenedores.  Navegue sin esfuerzo entre los procesos dentro de su contenedor para alojar sus contenedores en los que se ejecutan, organizados en tablas expandibles y clasificables.  Encuentre fácilmente el contenedor utilizando la mayor cantidad de CPU o memoria para un host o servicio determinado.

## Interactuar y administrar contenedores

<img src="imgs/terminals.png" width="200" alt="Launch a command line." align="right">

Interactúe directamente con sus contenedores: pause, reinicie y detenga los contenedores. Inicie una línea de comandos. Todo ello sin salir de la ventana del navegador de ámbito.

## Amplíe y personalice a través de plugins

Agregue detalles o interacciones personalizados para sus hosts, contenedores y/o procesos mediante la creación de complementos de Scope. O bien, simplemente elija entre algunos que otros ya han escrito en el GitHub [Complementos de weaveworks Scope](https://github.com/weaveworks-plugins/) organización.

## Quién utiliza Scope en la producción

*   [Ábside](https://apester.com/)
*   [Deepfence](https://deepfence.io) en [ThreatMapper](https://github.com/deepfence/ThreatMapper) y [ThreatStryker](https://deepfence.io/threatstryker/)
*   [MayaData](https://mayadata.io/) en [MayaOnline / MayaOnPrem](https://mayadata.io/products)
*   [Tejidos](https://www.weave.works/) en [Tejer la nube](https://cloud.weave.works)

Si deseas que se agregue tu nombre, comunícanoslo en Slack o envía un PR, por favor.

## <a name="getting-started"></a>Empezar

**Asegúrese de que su computadora esté detrás de un firewall que bloquea el puerto 4040** entonces

```console
sudo curl -L git.io/scope -o /usr/local/bin/scope
sudo chmod a+x /usr/local/bin/scope
scope launch
```

Este script descarga y ejecuta una imagen de ámbito reciente de Docker Hub.
Ahora, abra su navegador web para **<http://localhost:4040>**.

Para obtener instrucciones sobre cómo instalar Scope en [Kubernetes](https://www.weave.works/docs/scope/latest/installing/#k8s), [DCOS](https://www.weave.works/docs/scope/latest/installing/#dcos)o [ECS](https://www.weave.works/docs/scope/latest/installing/#ecs)ver [los documentos](https://www.weave.works/docs/scope/latest/introducing/).

## <a name="help"></a>Intenta comunicarte

Somos una comunidad muy amigable y nos encantan las preguntas, la ayuda y los comentarios.

Si tiene alguna pregunta, comentario o problema con Scope:

*   Docs
    *   Leer [los documentos de Weave Scope](https://www.weave.works/docs/scope/latest/introducing/)
    *   Echa un vistazo a la [preguntas frecuentes](/site/faq.md)
    *   Más información sobre cómo el [La comunidad scope opera](GOVERNANCE.md)
*   Únete a la discusión
    *   Invítate a la <a href="https://slack.weave.works/" target="_blank">Comunidad de tejidos</a> Flojo
    *   Haga una pregunta en el [#scope](https://weave-community.slack.com/messages/scope/) Canal de Slack
    *   Enviar un correo electrónico a [Grupo comunitario Scope](https://groups.google.com/forum/#!forum/scope-community)
*   Reuniones y eventos
    *   Únete a la [Grupo de usuarios de Tejido](https://www.meetup.com/pro/Weave/) y ser invitado a charlas en línea, capacitación práctica y reuniones en su área
    *   Únete (y sigue leyendo) a los regulares [Reuniones comunitarias de scope](https://docs.google.com/document/d/103\_60TuEkfkhz_h2krrPJH8QOx-vRnPpbcCZqrddE1s/edit) - actualmente en espera.
*   Contribuyendo
    *   Averigüe cómo [contribuir a Scope](CONTRIBUTING.md)
    *   [Presentar un problema](https://github.com/weaveworks/scope/issues/new) o hacer una solicitud de extracción para uno de nuestros [buenos primeros números](https://github.com/weaveworks/scope/issues?q=is%3Aissue+is%3Aopen+label%3Agood-first-issue)

¡Sus comentarios son siempre bienvenidos!

Seguimos el [Código de conducta de la CNCF](CODE-OF-CONDUCT.md).

## Licencia

Scope está licenciado bajo la Licencia Apache, Versión 2.0. Ver [LICENCIA](LICENSE) para el texto completo de la licencia.\
Encuentre más detalles sobre las licencias de código de proveedor en [VENDORED_CODE.md](VENDORED_CODE.md).
