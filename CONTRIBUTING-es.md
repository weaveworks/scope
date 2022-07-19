# Cómo contribuir

El alcance es [Licencia apache 2.0](LICENSE) y acepta contribuciones a través de GitHub
solicitudes de extracción. Este documento describe algunas de las convenciones sobre desarrollo
flujo de trabajo, formato de mensajes de confirmación, puntos de contacto y otros recursos para realizar
es más fácil conseguir que su contribución sea aceptada.

Agradecemos las mejoras en la documentación, así como en el código.

## Certificado de Origen

Al contribuir a este proyecto, usted acepta el Certificado de Desarrollador de
Origen (DCO). Este documento fue creado por la comunidad Linux Kernel y es un
simple declaración de que usted, como contribuyente, tiene el derecho legal de hacer el
contribución. No se requiere ninguna acción de su parte, pero es una buena idea ver el
[DCO](DCO) para obtener detalles antes de comenzar a contribuir con código a Scope.

## Correo electrónico, chat y reuniones comunitarias

El proyecto utiliza la lista de correo electrónico de la comunidad de alcance y Slack:

*   Correo electrónico: [ámbito-comunidad](https://groups.google.com/forum/#!forum/scope-community)
*   Chat: Únete al [Comunidad de tejidos](https://weaveworks.github.io/community-slack/) Espacio de trabajo de Slack y usa el [#scope](https://weave-community.slack.com/messages/scope/) canal

Al enviar correo electrónico, generalmente es mejor usar la lista de correo. Los mantenedores suelen estar bastante ocupados y la lista de correo encontrará más fácilmente a alguien que pueda responder rápidamente. También estarás ayudando potencialmente a otros que tenían la misma pregunta.

**Actualmente en espera**: También nos reunimos regularmente en el [Reunión de la comunidad scope](https://docs.google.com/document/d/103\_60TuEkfkhz_h2krrPJH8QOx-vRnPpbcCZqrddE1s/). No se sienta desanimado a asistir a la reunión debido a no ser un desarrollador. ¡Todos son bienvenidos!

Seguimos el [Código de conducta de la CNCF](CODE-OF-CONDUCT.md).

## Empezar

*   Bifurcar el repositorio en GitHub
*   Lea el [LÉAME](README.md) para comenzar como usuario y aprender cómo/dónde pedir ayuda
*   Si desea contribuir como desarrollador, continúe leyendo este documento para obtener más instrucciones.
*   ¡Juega con el proyecto, envía errores, envía solicitudes de extracción!

## Flujo de trabajo de contribución

Este es un esquema aproximado de cómo preparar una contribución:

*   Cree una rama temática desde la que desee basar su trabajo (generalmente ramificada desde el maestro).
*   Realizar confirmaciones de unidades lógicas.
*   Asegúrese de que sus mensajes de confirmación estén en el formato adecuado (consulte a continuación).
*   Inserte los cambios en una rama de temas en la bifurcación del repositorio.
*   Si cambiaste el código:
    *   Agregar pruebas automatizadas para cubrir los cambios
*   Envíe una solicitud de extracción al repositorio original.

## Cómo compilar y ejecutar el proyecto

```bash
make && ./scope stop && ./scope launch
```

Después de cada cambio que realice en el código Go, deberá volver a ejecutar el comando anterior (recompilar y reiniciar Scope) y actualizar la pestaña del navegador para ver sus cambios.

**Propina**: Si solo está realizando cambios en el código frontend de Scope, puede acelerar el ciclo de desarrollo iniciando adicionalmente el servidor Webpack, que recompilará y recargará automáticamente la pestaña de su navegador http://localhost:4042 en cada cambio:

```bash
cd client && yarn install && yart start
```

## Cómo ejecutar el conjunto de pruebas

### Backend

Puede ejecutar el linting go y las pruebas unitarias simplemente haciendo

```bash
make tests
```

Hay pruebas de integración para Scope, pero desafortunadamente es difícil configurarlas en repositorios bifurcados y la configuración no está documentada. Se necesita ayuda para mejorar esta situación: [#2192](https://github.com/weaveworks/scope/issues/2192)

### Frontend

Uso `yarn` para ejecutar todas las pruebas de Javascript y comprobaciones de linting:

```bash
cd client && yarn install && yarn test && yarn lint
```

## Política de aceptación

Estas cosas harán que un PR sea más probable que sea aceptado:

*   un requisito bien descrito
*   Pruebas de código nuevo
*   pruebas para código antiguo!
*   El nuevo código y las pruebas siguen las convenciones del código y las pruebas antiguos
*   un buen mensaje de confirmación (ver más abajo)

En general, fusionaremos un PR una vez que dos mantenedores lo hayan respaldado.
Los cambios triviales (por ejemplo, correcciones a la ortografía) pueden ser agitados.
Para cambios sustanciales, más personas pueden involucrarse, y es posible que se le pida que vuelva a presentar el PR o divida los cambios en más de un PR.

### Formato del mensaje de confirmación

Seguimos una convención aproximada para mensajes de confirmación que está diseñada para responder a dos
preguntas: qué cambió y por qué. La línea de asunto debe presentar el qué y
el cuerpo del compromiso debe describir el por qué.

```txt
scripts: add the test-cluster command

this uses tmux to setup a test cluster that you can easily kill and
start for debugging.

Fixes #38
```

El formato se puede describir de manera más formal de la siguiente manera:

```txt
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

La primera línea es el asunto y no debe tener más de 70 caracteres, el
la segunda línea siempre está en blanco, y otras líneas deben envolverse a 80 caracteres.
Esto permite que el mensaje sea más fácil de leer en GitHub, así como en varios
herramientas git.

## Plugins de 3ª parte

Así que has creado un plugin de Scope. ¿Dónde debería vivir?

Hasta que madure, debe vivir en su propio repositorio. Le recomendamos que anuncie su complemento en el [lista de correo](https://groups.google.com/forum/#!forum/scope-community) y para demostrarlo en un [reuniones comunitarias](https://docs.google.com/document/d/103\_60TuEkfkhz_h2krrPJH8QOx-vRnPpbcCZqrddE1s/).

Si tiene una buena razón por la cual los mantenedores de Scope deben tomar la custodia de su
plugin, por favor abra un problema para que pueda ser potencialmente promovido a la [Plugins de alcance](https://github.com/weaveworks-plugins/) organización.
