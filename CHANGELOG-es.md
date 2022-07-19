## Versión 1.13.2

Principalmente actualizaciones de dependencias, además de un par de pequeñas mejoras.

Gracias a todos los que contribuyeron a este lanzamiento: @gaby, @Kielek, @knrt10

### Mejoras

*   Establecer el nombre de host en el nombre del nodo de Kubernetes
    [#3827](https://github.com/weaveworks/scope/pull/3827)
*   Detener la representación si se cancela el contexto
    [#3801](https://github.com/weaveworks/scope/pull/3801)

### Correcciones

*   Solo montar el directorio de plugins si existe
    [#3825](https://github.com/weaveworks/scope/pull/3825)
*   Facturación multiinquilino: haga frente al intervalo de espionaje establecido durante más tiempo que el intervalo de publicación
    [#3796](https://github.com/weaveworks/scope/pull/3796)
*   Consulta multiinquilino: no elimines tanto las topologías
    [#3791](https://github.com/weaveworks/scope/pull/3791)
*   Multiinquilino: escanear líneas de comandos de contenedores y procesos
    [#3789](https://github.com/weaveworks/scope/pull/3789)

### Actualizaciones de dependencias

*   Actualizar la imagen base alpina a 3.13
    [#3838](https://github.com/weaveworks/scope/pull/3838)
*   Actualizar herramientas de compilación, incluida Go 1.16.2
    [#3833](https://github.com/weaveworks/scope/pull/3833)
    [#3797](https://github.com/weaveworks/scope/pull/3797)
    [#3821](https://github.com/weaveworks/scope/pull/3821)
*   actualizar lodash a 4.17.20 (CVE-2020-8203)
    [#3831](https://github.com/weaveworks/scope/pull/3831)
*   actualizar dot-prop, webpack y terser-webpack-plugin
    [#3816](https://github.com/weaveworks/scope/pull/3816)
*   bump http-proxy de 1.16.2 a 1.18.1 en /client
    [#3819](https://github.com/weaveworks/scope/pull/3819)
*   bump elíptico de 6.4.0 a 6.5.3 en /client
    [#3814](https://github.com/weaveworks/scope/pull/3814)
*   bump lodash de 4.17.15 a 4.17.19 en /client
    [#3812](https://github.com/weaveworks/scope/pull/3812)
*   Actualización a los weaveworks/ui-components más recientes
    [#3795](https://github.com/weaveworks/scope/pull/3795)
*   actualizar JS a 6.0.3
    [#3785](https://github.com/weaveworks/scope/pull/3785)

### Construir y probar

*   Ejecutar la AWS CLI desde una imagen de contenedor
    [#3841](https://github.com/weaveworks/scope/pull/3841)
*   Eliminar github.com/fatih/hclfmt de dependencia obsoletos
    [#3834](https://github.com/weaveworks/scope/pull/3834)

## Versión 1.13.1

Esta versión corrige un error en las operaciones de 'kubernetes describe', donde
La sonda giraría para volver a abrir una conexión a la interfaz de usuario una y otra vez
una vez finalizada la operación.

También elimina algunos códigos obsoletos que se conectaban al local no seguro
puerto kubelet en Kubernetes; Actualice la configuración a un
sonda única para hablar con Kubernetes para todo el clúster si
aún no lo han hecho. Gracias a @CiMaol por esta contribución.

### Correcciones

*   Detener las operaciones de 'kubernetes describe' girando
    [#3784](https://github.com/weaveworks/scope/pull/3784)
*   Sondeo: omitir la publicación de informes vacíos cuando la tasa de publicación es superior a la tasa de recopilación
    [#3774](https://github.com/weaveworks/scope/pull/3774)

### Mejoras

*   Probe ya no habla con kubelet local
    [#3754](https://github.com/weaveworks/scope/pull/3754)
*   Seguimiento del error de redondeo en el cálculo de facturación multiinquilino
    [#3779](https://github.com/weaveworks/scope/pull/3779)

### Rendimiento

*   Multiinquilino: combine los informes entrantes en el recopilador para ahorrar tiempo de E/S y consulta
    [#3780](https://github.com/weaveworks/scope/pull/3780)
    [#3781](https://github.com/weaveworks/scope/pull/3781)
    [#3782](https://github.com/weaveworks/scope/pull/3782)

### Dependencias

*   actualizar html-webpack-plugin a la estabilidad más reciente
    [#3776](https://github.com/weaveworks/scope/pull/3776)
*   Biblioteca Downgrade Fluent-Logger-Golang utilizada en modo multiinquilino
    [#3772](https://github.com/weaveworks/scope/pull/3772)

## Versión 1.13.0

Esta versión trae algunas correcciones de errores y número de rendimiento
mejoras, en particular en la reducción de los datos enviados cuando hay
muchas conexiones de socket entre dos puntos finales.

El aumento en el número de versión refleja un cambio en el protocolo de cable para esto
cambio en los datos de los extremos, y también un cambio en la forma en que los controles activos
están codificados.

Gracias a todos los que contribuyeron a este lanzamiento: @DarthSett,
@sarataha, @slalwani97 y @qiell.

### Correcciones

*   Quitar ceros finales en grandes cantidades en la interfaz de usuario
    [#3760](https://github.com/weaveworks/scope/pull/3760)
*   kubernetes: mostrar el estado del pod como "terminando" cuando corresponda
    [#3729](https://github.com/weaveworks/scope/pull/3729)
*   kubernetes: detectar más contenedores de 'pausa'
    [#3743](https://github.com/weaveworks/scope/pull/3743)
*   Mejore el cálculo del uso en código multiinquilino
    [#3751](https://github.com/weaveworks/scope/pull/3751)
    [#3753](https://github.com/weaveworks/scope/pull/3753)

### Mejoras de rendimiento

*   Elide muchas conexiones desde/hacia los mismos puntos finales
    [#3709](https://github.com/weaveworks/scope/pull/3709)
*   Eliminar dos estructuras de datos especializadas; unificar con otros datos de nodo
    [#3714](https://github.com/weaveworks/scope/pull/3714)
    [#3748](https://github.com/weaveworks/scope/pull/3748)
*   Simplifique algunos renderizadores para mejorar el rendimiento
    [#3747](https://github.com/weaveworks/scope/pull/3747)
*   Ralentizar el intervalo de sondeo de DNS para reducir la actividad de la red
    [#3758](https://github.com/weaveworks/scope/pull/3758)

### Mejoras menores

*   Agregar el encabezado "user-agent" a las llamadas http desde la sonda Scope
    [#3720](https://github.com/weaveworks/scope/pull/3720)
*   Establecer marca de tiempo y ventana en cada informe
    [#3752](https://github.com/weaveworks/scope/pull/3752)
*   Agregar seguimiento para operaciones de tubería
    [#3745](https://github.com/weaveworks/scope/pull/3745)

### Actualizaciones de dependencias

*   Módulos Convert to Go
    [#3742](https://github.com/weaveworks/scope/pull/3742)
*   Actualización a Go 1.13.9
    [#3766](https://github.com/weaveworks/scope/pull/3766)
*   Go: actualizar las dependencias de weaveworks, prometheus, protobuf, jaeger y aws
    [#3745](https://github.com/weaveworks/scope/pull/3745)
    [#3756](https://github.com/weaveworks/scope/pull/3756)
*   JavaScript: actualizar babel, jest, webpack y otras dependencias; dedupe yarn.lock
    [#3733](https://github.com/weaveworks/scope/pull/3733)
    [#3755](https://github.com/weaveworks/scope/pull/3755)
    [#3757](https://github.com/weaveworks/scope/pull/3757)
    [#3763](https://github.com/weaveworks/scope/pull/3763)

## Versión 1.12.0

### Resúmenes

Admite los tipos de objetos 'v1' de Kubernetes que se necesitan para Kubernetes
1.16 y elimina el soporte para tipos obsoletos 'v1beta'.
[#3691](https://github.com/weaveworks/scope/pull/3691)

El formato de serialización cambia: los datos DNS se renombraron accidentalmente
'nodos' en la versión 1.11.6, y esta versión lo cambia de nuevo a 'DNS'.
[#3713](https://github.com/weaveworks/scope/pull/3713)

Gracias a todos los que contribuyeron a este lanzamiento: @bensooraj,
@chandankumar4, @imazik, @oleggator y @qiell.

### Mejoras menores

*   Permitir al usuario deshabilitar los complementos a través de la bandera de línea de comandos.
    [#3703](https://github.com/weaveworks/scope/pull/3703)
*   En la interfaz de usuario, reemplace JSON.stringify por json-stable-stringify
    [#3701](https://github.com/weaveworks/scope/pull/3701)

### Errores y correcciones de seguridad

*   Corrección: informe del error HTTP si se produce un error en la llamada /api
    [#3702](https://github.com/weaveworks/scope/pull/3702)
*   Solucione un bloqueo raro en el rastreador de conexiones ebpf alimentando las conexiones iniciales de forma sincrónica al reiniciar.
    [#3712](https://github.com/weaveworks/scope/pull/3712)
*   Corregir errores tipográficos en cadenas de formato de depuración
    [#3695](https://github.com/weaveworks/scope/pull/3695)

### Mejoras de rendimiento

*   Manejar direcciones IP en binario en lugar de cadenas
    [#3696](https://github.com/weaveworks/scope/pull/3696)
*   En la aplicación multiinquilino, guarde E/S manteniendo los datos de actualización rápida fuera del almacén persistente.
    [#3716](https://github.com/weaveworks/scope/pull/3716)

### Actualizaciones de dependencias

*   Actualizar la versión go a 1.13.0
    [#3692](https://github.com/weaveworks/scope/pull/3692)
    [#3698](https://github.com/weaveworks/scope/pull/3698)
*   Actualizar la biblioteca de google/gopacket
    [#3606](https://github.com/weaveworks/scope/pull/3606)
*   Actualizar NodeJS a 8.12.0 y varias bibliotecas javascript
    [#3685](https://github.com/weaveworks/scope/pull/3685)
    [#3690](https://github.com/weaveworks/scope/pull/3690)
    [#3719](https://github.com/weaveworks/scope/pull/3719)
    [#3726](https://github.com/weaveworks/scope/pull/3726)

Mejoras en Build y CI:

*   Ejecute el contenedor de compilación de la interfaz de usuario como usuario actual para evitar que los archivos sean propiedad de la raíz.
    [#3635](https://github.com/weaveworks/scope/pull/3635)
*   Reemplazar archivos SASS con CSS y JavaScript
    [#3700](https://github.com/weaveworks/scope/pull/3700)
*   Eliminar el indicador -e obsoleto del inicio de sesión de Docker en CI
    [#3708](https://github.com/weaveworks/scope/pull/3708)
*   Corregir favicon.ico en modo de desarrollo de interfaz de usuario
    [#3705](https://github.com/weaveworks/scope/pull/3705)
*   Refactorizar la lectura del informe para simplificar el código
    [#3687](https://github.com/weaveworks/scope/pull/3687)
*   No importe fuentes cuando la interfaz de usuario de ámbito esté incrustada.
    [#3704](https://github.com/weaveworks/scope/pull/3704)

## Versión 1.11.6

Esta es en gran medida una versión de mejora del rendimiento: el mayor cambio
es que la investigación ahora publica informes completos una de cada tres veces; el
El resto son deltas que son mucho más pequeños, por lo tanto, utilizan menos CPU y memoria
en la aplicación. [#3677](https://github.com/weaveworks/scope/pull/3677)

También una nueva función de resumen de depuración en la aplicación, expuesta a través de http
[#3686](https://github.com/weaveworks/scope/pull/3686)

Algunas otras pequeñas mejoras de rendimiento:

*   perf(sonda): reducir la copia de nodos
    [#3679](https://github.com/weaveworks/scope/pull/3679)
*   perf(probe): agregue la etiqueta 'omitempty' a Topology.Nodes
    [#3678](https://github.com/weaveworks/scope/pull/3678)
*   perf(probe): actualice la biblioteca netlink para introducir mejoras en el rendimiento
    [#3681](https://github.com/weaveworks/scope/pull/3681)
*   perf(multitenant): cuantice la memoria caché de informes en el lado de consulta de aws-collector
    [#3671](https://github.com/weaveworks/scope/pull/3671)

Otros cambios:

*   Agregar tramos de seguimiento para la representación en la interfaz de usuario a través de websocket
    [#3682](https://github.com/weaveworks/scope/pull/3682)
*   Actualizar algunas dependencias de JavaScript
    [#3664](https://github.com/weaveworks/scope/pull/3664)
*   Actualiza los componentes de la interfaz de usuario a la versión con componentes con estilo 4
    [#3670](https://github.com/weaveworks/scope/pull/3670)
    [#3673](https://github.com/weaveworks/scope/pull/3673)
*   fix(test-flake): sondeo para el resultado en TestRegistryDelete() para evitar la carrera
    [#3688](https://github.com/weaveworks/scope/pull/3688)
*   refactorizar: quitar código de controles innecesario antiguo
    [#3680](https://github.com/weaveworks/scope/pull/3680)
*   Detener el paquete de procesamiento en función de la sonda
    [#3675](https://github.com/weaveworks/scope/pull/3675)
*   Quitar algunas constantes de cadena no utilizadas
    [#3674](https://github.com/weaveworks/scope/pull/3674)

## Versión 1.11.5

Algunas pequeñas mejoras:

*   Exponer las métricas de la sonda a Prometheus, si se proporciona una dirección http-listen
    [#3600](https://github.com/weaveworks/scope/pull/3600)
*   Reducir el uso de recursos de sondeo causado por fugas ocasionales en el informador de extremos de sondeo
    [#3661](https://github.com/weaveworks/scope/pull/3661)
*   Actualizar la herramienta de comprobación de JavaScript 'eslint' y resolver advertencias
    [#3643](https://github.com/weaveworks/scope/pull/3643)

## Versión 1.11.4

Esta versión contiene algunas correcciones, una de las cuales debería mejorar
uso de recursos en hosts que tienen muchas conexiones TCP.

*   Mejore el rastreador de conexión eBPF para reducir el número de veces
    se reinicia y vuelve a caer a un mecanismo menos eficiente.
    [#3653](https://github.com/weaveworks/scope/pull/3653)
*   Agregue el nombre del reportero a los registros de errores de la sonda. Gracias a @princerachit
    [#3363](https://github.com/weaveworks/scope/pull/3363)
*   Aplazar el registro de métricas hasta que lo necesitemos
    [#3605](https://github.com/weaveworks/scope/pull/3605)
*   Eliminar la métrica no utilizada SpyDuration
    [#3646](https://github.com/weaveworks/scope/pull/3646)
*   Quitar quay.io del script de lanzamiento
    [#3657](https://github.com/weaveworks/scope/pull/3657)

## Versión 1.11.3

Esta es una versión de corrección de errores, que debería mejorar algunos casos en los que
los informes se hacen cada vez más grandes con el tiempo debido a que Scope no ve
las conexiones se cierran.

*   Informar del error y reiniciar cuando algo salga mal en el seguimiento de la conexión
    [#3648](https://github.com/weaveworks/scope/pull/3648)

## Versión 1.11.2

Actualización menor:

*   Actualizaciones de las dependencias de JavaScript, donde se informaron problemas de seguridad:
    [#3633](https://github.com/weaveworks/scope/pull/3633)
*   Otra solución para que Scope no se dé cuenta cuando un contenedor ha sido destruido:
    [#3627](https://github.com/weaveworks/scope/pull/3627)

## Versión 1.11.1

Esta versión corrige un par de errores:

*   A veces, la sonda no puede eliminar su registro de un contenedor cuando se destruye
    [#3623](https://github.com/weaveworks/scope/pull/3623)
*   El modo de contraste no se reflejaba en la URL
    [#3617](https://github.com/weaveworks/scope/pull/3617)

## Versión 1.11.0

### Resúmenes

*   Agregar control de descripción a todos los recursos de Kubernetes [#3589](https://github.com/weaveworks/scope/pull/3589)
*   Mostrar trabajos de Kubernetes [#3609](https://github.com/weaveworks/scope/pull/3609)

Gracias a todos los que contribuyeron a este lanzamiento: @Deepak1100, @SaifRehman, @awolde, @bboreham, @bia, @dholbach, @fbarl, @foot, @guyfedwards, @jrryjcksn, @leavest, @n0rig, @najeal, @paulmorabito, @qiell, @qurname2, @rade, @satyamz, @shindanim, @tiriplicamihai, @tvvignesh, @xrgy.

¡Un agradecimiento especial a @qiell y @satyamz por implementar ambas características!

### Mejoras menores

*   Censura condicional de informes [#3571](https://github.com/weaveworks/scope/pull/3571)
*   Agregar un cuadro de diálogo de confirmación para eliminar un pod [#3572](https://github.com/weaveworks/scope/pull/3572)
*   Agregue openTracing span para el informe. ReadBinary() [#3598](https://github.com/weaveworks/scope/pull/3598)
*   Agregar métricas para el tamaño y el recuento de informes por inquilino [#3599](https://github.com/weaveworks/scope/pull/3599)

### Errores y correcciones de seguridad

*   Correcciones de seguridad en dependencias de desarrollo [#3578](https://github.com/weaveworks/scope/pull/3578)
*   Usar el contexto de viaje en el tiempo al descargar informes sin procesar [#3582](https://github.com/weaveworks/scope/pull/3582)
*   Corregir error Chrome 56+ no puede evitar los eventos predeterminados de la rueda [#3593](https://github.com/weaveworks/scope/pull/3593)
*   Reparar la sonda dnssnooper para varios CNAME [#3566](https://github.com/weaveworks/scope/pull/3566)
*   Corregir el script build-pkg [#3587](https://github.com/weaveworks/scope/pull/3587)

### Actualizaciones de dependencias

*   Actualizar a Webpack 4 [#3580](https://github.com/weaveworks/scope/pull/3580)
*   Actualizar la versión client-go a 10.0.0 [#3588](https://github.com/weaveworks/scope/pull/3588)
*   Actualizar componentes de interfaz de usuario y algunas de dependencias similares [#3574](https://github.com/weaveworks/scope/pull/3574)
*   Eliminar materialize-css JS dep [#3596](https://github.com/weaveworks/scope/pull/3596)

### Documentación y ejemplos

*   La ruta es en realidad examples/k8s [#3586](https://github.com/weaveworks/scope/pull/3586)
*   Ejemplos de actualización/permisos RBAC de k8s [#3595](https://github.com/weaveworks/scope/pull/3595)
*   Actualizar yamls de ejemplo para sondear Kubernetes una vez por clúster [#3569](https://github.com/weaveworks/scope/pull/3569)
*   Agregar solicitudes de CPU y memoria a manifiestos de Kubernetes de ejemplo [#3570](https://github.com/weaveworks/scope/pull/3570)
*   Amplíe las preguntas frecuentes con autenticación básica a través de variables de entorno [#3575](https://github.com/weaveworks/scope/pull/3575)

## Versión 1.10.2

Esta versión tiene una corrección de seguridad, algunas correcciones de errores y algunas otras correcciones menores
Mejoras.

Un agradecimiento especial a @arnulfojr, @Akashtic, @AVRahul, @carlosedp, @ycao56
para las contribuciones de la comunidad!

Seguridad y corrección de errores:

*   Solucionar la vulnerabilidad (CVE-2018-16487) al actualizar la dependencia lodash
    [#3568](https://github.com/weaveworks/scope/pull/3568)
*   Solucionar la vulnerabilidad (CVE-2019-0542) al actualizar la dependencia xterm.js
    [#3557](https://github.com/weaveworks/scope/pull/3557)
*   Corregir los nodos que faltan en el modo de clúster de Kubernetes
    [#3444](https://github.com/weaveworks/scope/pull/3444)
*   Mostrar todas las conexiones cuando un pod tiene varios PVC
    [#3553](https://github.com/weaveworks/scope/pull/3553)
*   Arreglar la conexión espuria a PVC con el mismo nombre en un espacio de nombres diferente
    [#3530](https://github.com/weaveworks/scope/pull/3530)
*   Corregir la vinculación del gráfico de métricas a la URL de supervisión externa
    [#3534](https://github.com/weaveworks/scope/pull/3534)
*   Corregir la ruta de montaje de compilación en las pruebas unitarias de la interfaz de usuario
    [#3558](https://github.com/weaveworks/scope/pull/3558)
*   Agregar ID interno adicional en los botones de control
    [#3565](https://github.com/weaveworks/scope/pull/3565)

Mejoras en las características:

*   Expanda dinámicamente la lista de espacios de nombres si pasa el cursor sobre ella
    [#3117](https://github.com/weaveworks/scope/pull/3117)
    [#3562](https://github.com/weaveworks/scope/pull/3562)
*   Hacer que el número de uso de memoria del contenedor de Scope coincida con el de Docker
    [#3435](https://github.com/weaveworks/scope/pull/3435)

Mejoras en la compilación:

*   Agregar compilación de ARM64
    [#3537](https://github.com/weaveworks/scope/pull/3537)
*   Agregue una prueba de CI para detectar el problema que provocó el error de la versión 1.10.0
    [#3440](https://github.com/weaveworks/scope/pull/3440)
*   Actualizar algunos componentes de la interfaz de usuario de 3ª parte
    [#3450](https://github.com/weaveworks/scope/pull/3450)
*   Reordenar algunos códigos para que sean más legibles
    [#3551](https://github.com/weaveworks/scope/pull/3551)
*   Pequeña mejora del rendimiento en la representación de notificaciones de volumen de Kubernetes
    [#3445](https://github.com/weaveworks/scope/pull/3445)

Mejoras en la documentación:

*   [#3417](https://github.com/weaveworks/scope/pull/3417)
    [#3447](https://github.com/weaveworks/scope/pull/3447)
    [#3448](https://github.com/weaveworks/scope/pull/3448)
    [#3436](https://github.com/weaveworks/scope/pull/3436)
    [#3545](https://github.com/weaveworks/scope/pull/3545)
    [#3546](https://github.com/weaveworks/scope/pull/3546)
    [#3563](https://github.com/weaveworks/scope/pull/3563)

## Versión 1.10.1

Este es un relanzamiento de 1.10.0 que fue golpeado por una desafortunada compilación
error.

*   Compilación de la interfaz de usuario: deje de eliminar archivos de interfaz de usuario estáticos al crear una interfaz de usuario externa
    [#3439](https://github.com/weaveworks/scope/pull/3439)

## Versión 1.10.0

Resúmenes:

*   Agregar instantáneas de volumen persistente de Kubernetes y operaciones de clonación
    [#3355](https://github.com/weaveworks/scope/pull/3355)

*   Los objetos de Kubernetes se pueden informar solo una vez en un clúster, en su lugar
    de reportar los mismos datos de cada nodo.
    [#3274](https://github.com/weaveworks/scope/pull/3274)
    [#3419](https://github.com/weaveworks/scope/pull/3419)
    [#3432](https://github.com/weaveworks/scope/pull/3432)

*   La aplicación ahora es compatible con http autenticación básica
    [#3393](https://github.com/weaveworks/scope/pull/3393)

Se realizaron algunos cambios (#3266, #3272) en el protocolo de cable, que
significa que las sondas nuevas no son compatibles con una aplicación anterior.

Gracias por las contribuciones de @Akash4927, @akshatnitd, @bhavin192,
@hexmind, @gfeun, @gotjosh, @gruebel, @hexmind, @jgsqware, @ltachet,
@muthumalla, @rvrvrv, @satyamz, @ScottBrenner, @ssiddhantsharma,
@visualapps, @WhiteHatTux, @ycao56, @Xivolkar, algunos de estos llegaron a través de
[Hacktoberfest](https://hacktoberfest.digitalocean.com/).

Rendimiento:

*   Sondeo: use netlink para hablar con conntrack
    [#3298](https://github.com/weaveworks/scope/pull/3298)
*   Quitar los miembros de datos Primero y Último de las estructuras de métricas
    [#3266](https://github.com/weaveworks/scope/pull/3266)
*   Eliminar el antiguo campo 'Controles' que fue reemplazado hace dos años
    [#3272](https://github.com/weaveworks/scope/pull/3272)
*   Sondeo: No informe de procesos muertos o difuntos
    [#3379](https://github.com/weaveworks/scope/pull/3379)
*   Simplifique la obtención de direcciones IP en un espacio de nombres
    [#3335](https://github.com/weaveworks/scope/pull/3335)
*   Descartar actualizaciones de pod para otros nodos
    [#3391](https://github.com/weaveworks/scope/pull/3391)
*   Sondeo: Publicación de informes de límite de velocidad
    [#3386](https://github.com/weaveworks/scope/pull/3386)
*   En la aplicación multiinquilino, suelte todos los nodos para topologías grandes
    [#3384](https://github.com/weaveworks/scope/pull/3384)

Correcciones de errores y mejoras menores:

*   Compatibilidad con la interfaz de tiempo de ejecución de contenedor inicial (CRI)
    [#3275](https://github.com/weaveworks/scope/pull/3275)
    [#3305](https://github.com/weaveworks/scope/pull/3305)
    [#3308](https://github.com/weaveworks/scope/pull/3308)
    [#3392](https://github.com/weaveworks/scope/pull/3392)
    [#3364](https://github.com/weaveworks/scope/pull/3364)
*   Agregar el nombre del controlador de almacenamiento al volumen persistente
    [#3260](https://github.com/weaveworks/scope/pull/3260)
*   Fix WithLatests() fixup en claves duplicadas
    [#3281](https://github.com/weaveworks/scope/pull/3281)
*   Agregar variante EKS de 'contenedor de pausa'
    [#3421](https://github.com/weaveworks/scope/pull/3421)
*   Agregar seguimiento distribuido de Opentracing (Jaeger) para crear perfiles de la aplicación
    [#3307](https://github.com/weaveworks/scope/pull/3307)
    [#3380](https://github.com/weaveworks/scope/pull/3380)
    [#3383](https://github.com/weaveworks/scope/pull/3383)
    [#3325](https://github.com/weaveworks/scope/pull/3325)
*   aplicación: actualizar mensaje de contenedor detenido
    [#3396](https://github.com/weaveworks/scope/pull/3396)
*   Ejemplo de archivos yaml de Kubernetes: corregir errores tipográficos y trabajar con Kubernetes más recientes
    [#3403](https://github.com/weaveworks/scope/pull/3403)
*   Ejemplo de archivos yaml de Kubernetes: agregar compatibilidad con PodSecurityPolicy
    [#3354](https://github.com/weaveworks/scope/pull/3354)
*   Ejemplo de archivos yaml de Kubernetes: agregar reglas en el rol de clúster para componentes de almacenamiento
    [#3290](https://github.com/weaveworks/scope/pull/3290)
*   Cambie el nombre de 'Hoja de almacenamiento' a 'Hoja' en los informes
    [#3323](https://github.com/weaveworks/scope/pull/3323)
    [#3324](https://github.com/weaveworks/scope/pull/3324)
*   Comprobar que el contenedor se está ejecutando antes de intentar abrir su espacio de nombres
    [#3279](https://github.com/weaveworks/scope/pull/3279)

Interfaz de usuario

*   Actualice a font-awesome 5 y nuevos iconos
    [#3426](https://github.com/weaveworks/scope/pull/3426)
*   Reemplazar el icono de compartir con el mapa del sitio en el botón gráfico
    [#3387](https://github.com/weaveworks/scope/pull/3387)
*   Versión de Bump ui-components
    [#3282](https://github.com/weaveworks/scope/pull/3282)
    [#3431](https://github.com/weaveworks/scope/pull/3431)
*   Usar el componente GraphNode de la biblioteca ui-components
    [#3262](https://github.com/weaveworks/scope/pull/3262)
*   Hacer que el encabezado sea semitransparente
    [#3294](https://github.com/weaveworks/scope/pull/3294)
*   Corregir el estilo roto del terminal en modo de contraste
    [#3347](https://github.com/weaveworks/scope/pull/3347)
*   Agregar enlaceRouteChange a la aplicación Ámbito
    [#3349](https://github.com/weaveworks/scope/pull/3349)
*   Impedir que dos instancias de ámbito en el mismo dominio cambien el historial de la otra
    [#3326](https://github.com/weaveworks/scope/pull/3326)
*   Usar el nuevo componente de búsqueda del repositorio ui-components
    [#3337](https://github.com/weaveworks/scope/pull/3337)
*   Actualizar localStorage con el estado Scope también en el enlace inicial del router
    [#3315](https://github.com/weaveworks/scope/pull/3315)

Cree y pruebe mejoras

*   Eliminar weaveutil y weave de Dockerfile.cloud-agent
    [#3369](https://github.com/weaveworks/scope/pull/3369)
*   Construya sobre la arquitectura power CPU.
    [#3231](https://github.com/weaveworks/scope/pull/3231)
*   build: Corregir la importación para golint que se ha movido
    [#3389](https://github.com/weaveworks/scope/pull/3389)
*   Suspender para detener el error TestRegistryDelete()
    [#3334](https://github.com/weaveworks/scope/pull/3334)
*   Arreglar el proveedor de ugorji / go
    [#3280](https://github.com/weaveworks/scope/pull/3280)
*   Actualizar la versión de sirupsen/logrus
    [#3276](https://github.com/weaveworks/scope/pull/3276)
    [#3277](https://github.com/weaveworks/scope/pull/3277)
*   Actualizar la versión client-go de Kubernetes a la 8.0.0
    [#3329](https://github.com/weaveworks/scope/pull/3329)
*   Actualizar la dependencia de lodash para eliminar la advertencia de seguridad
    [#3310](https://github.com/weaveworks/scope/pull/3310)
*   Reelaborar la compilación de la interfaz de usuario para mejorar el almacenamiento en caché y solucionar el problema de empaquetado
    [#3353](https://github.com/weaveworks/scope/pull/3353)
    [#3356](https://github.com/weaveworks/scope/pull/3356)
    [#3360](https://github.com/weaveworks/scope/pull/3360)
    [#3382](https://github.com/weaveworks/scope/pull/3382)
*   Actualizar el subdirectorio 'tools'
    [#3311](https://github.com/weaveworks/scope/pull/3311)
    [#3312](https://github.com/weaveworks/scope/pull/3312)
*   Limpiar Dockerfiles
    [#3411](https://github.com/weaveworks/scope/pull/3411)
*   Actualizar la copia suministrada de TCPTRACER-BPF por motivos de licencia
    [#3336](https://github.com/weaveworks/scope/pull/3336)
*   proveedor: actualizar gopkg.in/yaml.v2 a la última versión ascendente
    [#3317](https://github.com/weaveworks/scope/pull/3317)
*   Crear un archivo de parada bpf de forma diferente, en la prueba de integración
    [#3332](https://github.com/weaveworks/scope/pull/3332)
*   Pasar a CircleCI 2.0
    [#3333](https://github.com/weaveworks/scope/pull/3333)

## Versión 1.9.1

Resúmenes:

Scope ahora muestra Kubernetes Storage (PersistentVolume y
PersistentVolumeClaim) información en la vista Pods.
[#3132](https://github.com/weaveworks/scope/pull/3132)

¡Gracias a @satyamz y a todos en OpenEBS por esta contribución!

También gracias por el ejemplo kubernetes manifiestos de @tasdikrahman.

Correcciones de errores y mejoras menores:

*   Corregir los nodos 'No administrados' que se muestran a pesar del filtro 'Ocultar administrado'
    [#3189](https://github.com/weaveworks/scope/pull/3189)
*   Corrige la superposición de fuentes monoespaciales en terminal+linux
    [#3248](https://github.com/weaveworks/scope/pull/3248)
*   Hacer que la topología de proceso por nombre muestre algo
    [#3208](https://github.com/weaveworks/scope/pull/3208)
*   Usar el valor predeterminado para una topologíaOpción si se omite
    [#3165](https://github.com/weaveworks/scope/pull/3165)
*   Estimación ajustada de ancho/alto de caracteres de terminal
    [#3179](https://github.com/weaveworks/scope/pull/3179)
*   Corregir la detección de imágenes en pausa para Kubernetes 1.10
    [#3183](https://github.com/weaveworks/scope/pull/3183)
*   ebpf: comprobación de actualización para núcleos de Ubuntu defectuosos conocidos
    [#3188](https://github.com/weaveworks/scope/pull/3188)
*   Agregar opción para imprimir informes de sondeo en stdout, para depurar
    [#3204](https://github.com/weaveworks/scope/pull/3204)
*   Arreglar el pánico querier introducido en #3143
    [#3156](https://github.com/weaveworks/scope/pull/3156)
*   Corregir fallos raros en la función de filtro
    [#3232](https://github.com/weaveworks/scope/pull/3232)
*   Sondeo: corregir mensaje de error para nombrar el indicador correcto probe.proc.spy
    [#3216](https://github.com/weaveworks/scope/pull/3216)
*   Retire ProcessWithContainerNameRenderer, no estaba funcionando
    [#3263](https://github.com/weaveworks/scope/pull/3263)
*   Cierre la ventana del terminal a la salida; Actualizar xterm a la versión 3.3.0
    [#3172](https://github.com/weaveworks/scope/pull/3172)
*   Agregar etiquetas org.opencontainers.image.\* a Dockerfiles
    [#3171](https://github.com/weaveworks/scope/pull/3171)
*   Hacer que el encabezado de la tabla se alinee con columnas cuando aparezca la barra de desplazamiento
    [#3169](https://github.com/weaveworks/scope/pull/3169)
*   Agregar indicador de línea de comandos para establecer el tiempo de espera de SQS RPC
    [#3157](https://github.com/weaveworks/scope/pull/3157)

Rendimiento:

Una serie de pequeñas mejoras de rendimiento se han introducido en esto.
liberación, reduciendo el uso de memoria y CPU.

*   Sondeo: quitar código de compatibilidad con versiones anteriores al publicar informes
    [#3215](https://github.com/weaveworks/scope/pull/3215)
*   Optimizar Node.WithLatests()
    [#3268](https://github.com/weaveworks/scope/pull/3268)
*   Optimizar ConPadents() cuando solo hay un padre
    [#3269](https://github.com/weaveworks/scope/pull/3269)
*   Reutilizar escritores gzip en una piscina
    [#3267](https://github.com/weaveworks/scope/pull/3267)
*   Optimice la combinación donde un lado es un subconjunto del otro
    [#3253](https://github.com/weaveworks/scope/pull/3253)
*   Use un grupo de búferes en el informe. ReadBinary() para reducir la recolección de basura
    [#3255](https://github.com/weaveworks/scope/pull/3255)
*   Fusión de informes más rápida a través de objetos mutantes
    [#3236](https://github.com/weaveworks/scope/pull/3236)
*   Ruta más rápida para comprobar una dirección IP con redes conocidas
    [#3142](https://github.com/weaveworks/scope/pull/3142)
*   Omitir pods sin direcciones IP al representar conexiones de red
    [#3201](https://github.com/weaveworks/scope/pull/3201)
*   Obtener direcciones IP de contenedor directamente desde el espacio de nombres en lugar de llamar a 'weave ps'
    [#3207](https://github.com/weaveworks/scope/pull/3207)

HIG

Una serie de cambios ajustando las fuentes y los colores, y estandarizando el
INTERFAZ de usuario a través del uso de un tema.

*   Actualizar fuentes: use Proxima Nova como fuente predeterminada en lugar de Roboto.
    [#3177](https://github.com/weaveworks/scope/pull/3177)
*   Ajustar el tamaño de fuente
    [#3181](https://github.com/weaveworks/scope/pull/3181)
*   Actualizar los colores del tema gris
    [#3234](https://github.com/weaveworks/scope/pull/3234)
*   Usar nuevos colores de tema de acento
    [#3230](https://github.com/weaveworks/scope/pull/3230)
*   Usar nuevos colores de tema púrpura
    [#3229](https://github.com/weaveworks/scope/pull/3229)
*   Usar nuevos colores grises del tema
    [#3227](https://github.com/weaveworks/scope/pull/3227)
*   Dejar de usar colores de tema eliminados
    [#3148](https://github.com/weaveworks/scope/pull/3148)
*   Combinar colores de tema neutros
    [#3146](https://github.com/weaveworks/scope/pull/3146)
*   Fondo ligeramente aclarado para que coincida con el resto de WeaveCloud
    [#3206](https://github.com/weaveworks/scope/pull/3206)
*   Texto con mayúsculas y minúsculas en todas partes
    [#3166](https://github.com/weaveworks/scope/pull/3166)
*   Mostrar la etiqueta de imagen más claramente en los detalles del nodo
    [#3173](https://github.com/weaveworks/scope/pull/3173)
*   Estandarizar el radio de borde
    [#3170](https://github.com/weaveworks/scope/pull/3170)
*   Aplicar tamaños de fuente de tema
    [#3167](https://github.com/weaveworks/scope/pull/3167)
*   Usar solo valores de índice z del tema
    [#3159](https://github.com/weaveworks/scope/pull/3159)

Tejer la nube específica

Así como algunas correcciones de errores, refactorización de lugares donde se encuentra la integración
de Scope en la interfaz de usuario alojada de Weave Cloud complicó el código.

*   Uso correcto de api.getFluxImages
    [#3233](https://github.com/weaveworks/scope/pull/3233)
*   Mostrar implementaciones en Viajes en el tiempo
    [#3222](https://github.com/weaveworks/scope/pull/3222)
*   Separar el espacio de nombres del extremo de la API de la parte de ruta de url
    [#3221](https://github.com/weaveworks/scope/pull/3221)
*   Corregir la URL de descarga del informe de alcance en Weave Cloud
    [#3213](https://github.com/weaveworks/scope/pull/3213)
*   Usar el componente TimestampTag común
    [#3195](https://github.com/weaveworks/scope/pull/3195)
*   Cambiar la resolución de URL para acomodar las rutas de Weave Cloud
    [#3175](https://github.com/weaveworks/scope/pull/3175)
*   Soporte de representación de detalles de nodo extras
    [#3244](https://github.com/weaveworks/scope/pull/3244)
*   Inyección de TimeTravel de soporte
    [#3239](https://github.com/weaveworks/scope/pull/3239)

## Versión 1.9.0

Resúmenes:

*   Cambio en el comportamiento de los datos de la tabla: las etiquetas de Docker ahora se envían
    full, mientras que las variables de entorno de Docker no se notifican de forma predeterminada
*   Los plugins ahora pueden representar enlaces http y mostrar controles en más objetos

Nuevas características del plugin:

*   Representar vínculos http en tablas
    [#3105](https://github.com/weaveworks/scope/pull/3105)
*   Soporta controles de plugins en K8s Service, DaemonSet, StatefulSet, Cronjob.
    [#3110](https://github.com/weaveworks/scope/pull/3110)

Correcciones de errores y mejoras menores:

*   Evitar el bloqueo del kernel de Ubuntu
    [#3141](https://github.com/weaveworks/scope/pull/3141)
*   Dejar de truncar tablas; Deshabilitar los informes de Vars de Docker de forma predeterminada
    [#3139](https://github.com/weaveworks/scope/pull/3139)
*   No mostrar pods fallidos
    [#3126](https://github.com/weaveworks/scope/pull/3126)
*   Haga que el ámbito comience con Docker para Mac nuevamente.
    [#3140](https://github.com/weaveworks/scope/pull/3140)
*   Corregir el historial del navegador al vincular profundamente los detalles del nodo con el contexto de tiempo
    [#3134](https://github.com/weaveworks/scope/pull/3134)
*   Pasar a un tema de color más consistente
    [#3116](https://github.com/weaveworks/scope/pull/3116)
    [#3124](https://github.com/weaveworks/scope/pull/3124)
    [#3136](https://github.com/weaveworks/scope/pull/3136)
*   Cadena de formato de corrección que solo se usa en la depuración
    [#3129](https://github.com/weaveworks/scope/pull/3129)
*   Corregir documentos para la instalación de OpenShift
    [#3128](https://github.com/weaveworks/scope/pull/3128)

Rendimiento:

*   Usar combinación no segura en joinResults.addChildAndChildren()
    [#3143](https://github.com/weaveworks/scope/pull/3143)
*   Usar ruta de código de un solo propietario para acumular elementos secundarios al representar
    [#3138](https://github.com/weaveworks/scope/pull/3138)
*   Simplificar Map.Render()
    [#3135](https://github.com/weaveworks/scope/pull/3135)
*   Permita que la sonda envíe informes de "acceso directo" más pequeños para actualizar la interfaz de usuario más rápido
    [#3121](https://github.com/weaveworks/scope/pull/3121)

## Versión 1.8.0

Resúmenes:

*   Muchas mejoras de rendimiento
*   Un cambio en el protocolo de cable (ver #3061 a continuación - la nueva aplicación es
    compatible con sondas más antiguas pero no al revés)

Nuevas características y mejoras:

*   Agregar el tipo de servicio y los puertos de Kubernetes a la pantalla Servicios
    [#3090](https://github.com/weaveworks/scope/pull/3090)

Correcciones de errores y mejoras menores:

*   Renovar las instrucciones de instalación
    [#3052](https://github.com/weaveworks/scope/pull/3052)
*   Corregir los nodos 'No administrados' que se muestran a pesar del filtro 'Ocultar administrado'
    [#3097](https://github.com/weaveworks/scope/pull/3097)
*   Eliminar grandes espacios entre el encabezado y la tabla
    [#3066](https://github.com/weaveworks/scope/pull/3066)
*   Valor en blanco en la inserción de decodificación de LatestMap
    [#3095](https://github.com/weaveworks/scope/pull/3095)
*   refactorizar: no devuelva el receptor en Topology.AddNode()
    [#3075](https://github.com/weaveworks/scope/pull/3075)
*   Quitar la función de árbol de procesos no utilizada GetChildren()
    [#3094](https://github.com/weaveworks/scope/pull/3094)

Mejoras de rendimiento:

*   Mover la asignación de nombres DNS del extremo al informe
    [#3061](https://github.com/weaveworks/scope/pull/3061)
*   Habilitar la configuración para dejar de solicitar la lista de pods de kubelet, a través de la variable de entorno
    [#3077](https://github.com/weaveworks/scope/pull/3077)
*   Excluir entradas nulas para redes en nodos contenedores en el informe de sondeo
    [#3091](https://github.com/weaveworks/scope/pull/3091)
*   Elimine el indicador -probe.kubernetes.interval y deje de volver a sincronizar los datos de Kubernetes
    [#3080](https://github.com/weaveworks/scope/pull/3080)
*   Optimizar procesoTopología()
    [#3074](https://github.com/weaveworks/scope/pull/3074)
*   Etiquetador docker más eficiente
    [#3093](https://github.com/weaveworks/scope/pull/3093)
*   Agregar topología. ReplaceNode() para mayor eficiencia
    [#3073](https://github.com/weaveworks/scope/pull/3073)
*   Establecer 'omitempty' en Node Adjacency
    [#3062](https://github.com/weaveworks/scope/pull/3062)

Seguridad:

*   Aumentar las dependencias de JavaScript para recoger la solución para el asesoramiento de seguridad
    [#3102](https://github.com/weaveworks/scope/pull/3102)

Construir y probar:

*   Agregar una prueba que compruebe si los informes con datos de ida y vuelta
    [#2399](https://github.com/weaveworks/scope/pull/2399)
*   Guarde el código fuente generado como un artefacto de CI, en caso de que sea necesario
    para la solución de problemas.
    [#3056](https://github.com/weaveworks/scope/pull/3056)

Cambios relacionados con Weave Cloud:

*   Deshabilite el enlace del panel de detalles si la supervisión no está disponible.
    [#3070](https://github.com/weaveworks/scope/pull/3070)
    [#3072](https://github.com/weaveworks/scope/pull/3072)
*   Agregar (nube).) weave.works a la lista de servicios conocidos
    [#3084](https://github.com/weaveworks/scope/pull/3084)
*   Modificar solo el título del documento si se ejecuta de forma independiente
    [#3071](https://github.com/weaveworks/scope/pull/3071)
*   Corrige un error que muestra el "estado de la imagen del contenedor" en todos los tipos de recursos
    [#3054](https://github.com/weaveworks/scope/pull/3054)
*   Cambios relacionados con las visitas guiadas
    [#3068](https://github.com/weaveworks/scope/pull/3068)
    [#3088](https://github.com/weaveworks/scope/pull/3088)
*   Mostrar el viaje en el tiempo en todo momento en Weave Cloud
    [#3065](https://github.com/weaveworks/scope/pull/3065)
*   Enlace de cpu/mem del servicio de actualización
    [#3060](https://github.com/weaveworks/scope/pull/3060)

## Versión 1.7.3

Correcciones de errores y mejoras menores:

*   Corrige el problema en el que, si el servidor de API era inaccesible en el inicio, no se informaría de ningún recurso de kubernetes.
    [#3050](https://github.com/weaveworks/scope/pull/3050)

## Versión 1.7.2

Resúmenes:

*   Rastreador eBPF que funciona en GKE: esto hace que el seguimiento de la conexión sea más eficiente y preciso

Nuevas características y mejoras:

*   proveedor: bump tcptracer-bpf
    [#3042](https://github.com/weaveworks/scope/pull/3042)

Correcciones de errores y mejoras menores:

*   Cierre la tubería del terminal, al cerrar el panel de la cápsula
    [#3045](https://github.com/weaveworks/scope/pull/3045)
*   Obtenga cronjobs de 'batch/v1beta1'. Esto corrige un error que causaba que CronJobs en k8s recientes no apareciera en Scope.
    [#3044](https://github.com/weaveworks/scope/pull/3044)

Documentación:

*   Actualizar las instrucciones de instalación para usar el espacio de nombres weave
    [#3041](https://github.com/weaveworks/scope/pull/3041)

## Versión 1.7.1

Resúmenes:

*   Se introdujo un error en 1.7.0 al cerrar el panel de terminales de registro de pod que hace que la sonda gire,
    por lo tanto, saturando una cpu. Esto se ha corregido en #3034.
*   Solucionar un problema que haría que la sonda no informara de ciertos recursos de kubernetes si, al iniciarse,
    no pudo conectarse correctamente a la API de kubernetes.

Correcciones de errores y mejoras menores:

*   logReadCloser: asegúrese de que EOF después `Close()`
    [#3034](https://github.com/weaveworks/scope/pull/3034)
*   Compruebe si los recursos de k8s son compatibles con `runReflectorUntil`
    [#3037](https://github.com/weaveworks/scope/pull/3037)
*   Detener el enrutador de página al desmontar la aplicación
    [#3025](https://github.com/weaveworks/scope/pull/3025)
*   cliente: Corregir la ordenación del tiempo de actividad en la vista de tabla
    [#3038](https://github.com/weaveworks/scope/pull/3038)

Mejoras internas y limpieza:

*   Quitar valores predeterminados del hash de estado de URL
    [#3030](https://github.com/weaveworks/scope/pull/3030)
*   Manejar correctamente la reanudación del viaje en el tiempo en monitor
    [#3028](https://github.com/weaveworks/scope/pull/3028)
*   Cambiar el formato pausedAt de moment() a cadena ISO
    [#3036](https://github.com/weaveworks/scope/pull/3036)

## Versión 1.7.0

Resúmenes:

*   La visualización de los registros de pod ahora muestra todos los registros de contenedor con cada línea prefijada por `[containerName]`.
    Anteriormente, la misma vista fallaba si el pod tenía varios contenedores.
*   Mostrar todos los espacios de nombres de Kubernetes, incluidos los vacíos.
*   Varias pequeñas mejoras y trabajo de rendimiento

Nuevas características y mejoras:

*   La lectura de registros de pod devuelve todos los registros de contenedor
    [#3013](https://github.com/weaveworks/scope/pull/3013)
*   Espacios de nombres de informes de sondeo
    [#2985](https://github.com/weaveworks/scope/pull/2985)
*   mostrar procesos no conectados
    [#3009](https://github.com/weaveworks/scope/pull/3009)

Correcciones de errores y mejoras menores:

*   Establecer un tiempo de espera en animación de terminal
    [#3021](https://github.com/weaveworks/scope/pull/3021)
*   'updateKubeFilters' regresa antes de tiempo si no hay espacios de nombres
    [#3017](https://github.com/weaveworks/scope/pull/3017)
*   No asigne adyacencias de imágenes a hosts
    [#2997](https://github.com/weaveworks/scope/pull/2997)
*   hacer frente a asignaciones de topología de un >muchos
    [#2996](https://github.com/weaveworks/scope/pull/2996)
*   Etiquetar imágenes en tiempo de compilación
    [#2987](https://github.com/weaveworks/scope/pull/2987)
*   no excluya las conexiones NATed en la asignación a procesos
    [#2978](https://github.com/weaveworks/scope/pull/2978)

Mejoras internas y limpieza:

*   refactorizar: extraer código común en la asignación de extremos
    [#3016](https://github.com/weaveworks/scope/pull/3016)
*   refactorizar: hacer de PropagateSingleMetrics un renderizador
    [#3008](https://github.com/weaveworks/scope/pull/3008)
*   refactorizar: mover RenderContext donde pertenece
    [#3005](https://github.com/weaveworks/scope/pull/3005)
*   Representar etiquetas sensibles para nodos con pocos o ningún metadato
    [#2998](https://github.com/weaveworks/scope/pull/2998)
*   refactorizar: desterrar TheInternet
    [#3003](https://github.com/weaveworks/scope/pull/3003)
*   resumen del informe de referencia
    [#3000](https://github.com/weaveworks/scope/pull/3000)
*   refactorización: resumen en línea de metadatos, métricas, tablas
    [#2999](https://github.com/weaveworks/scope/pull/2999)
*   Actualizar Ir a 1.9.2
    [#2993](https://github.com/weaveworks/scope/pull/2993)
*   simplificar `joinResults`
    [#2994](https://github.com/weaveworks/scope/pull/2994)
*   Sugerir cómo deshabilitar los errores y advertencias de tejido
    [#2990](https://github.com/weaveworks/scope/pull/2990)
*   refactorizar: soltar redes desde render. MapFunc
    [#2991](https://github.com/weaveworks/scope/pull/2991)

Mejoras de rendimiento:

*   quitar Node.Edges
    [#2992](https://github.com/weaveworks/scope/pull/2992)
*   eliminar la propagación innecesaria de metadatos
    [#3007](https://github.com/weaveworks/scope/pull/3007)
*   configuración de permisos `probe.kubernetes.interval` a 0
    [#3012](https://github.com/weaveworks/scope/pull/3012)
*   Detener la captura de ReplicaSets y ReplicationControllers
    [#3014](https://github.com/weaveworks/scope/pull/3014)
*   optimización: asignación previa y menos segmentos durante el resumen
    [#3002](https://github.com/weaveworks/scope/pull/3002)
*   hacer `Report.Topology(name)` rápido
    [#3001](https://github.com/weaveworks/scope/pull/3001)

Cambios relacionados con Weave Cloud:

*   Bump ui-components a v0.4.18
    [#3019](https://github.com/weaveworks/scope/pull/3019)
*   Simplificación de los fondos para que coincidan con la grisada en service-ui y ui-compo...
    [#3011](https://github.com/weaveworks/scope/pull/3011)

## Versión 1.6.7

Esta es una versión de parche menor.

Mejoras internas y limpieza:

*   Actualizar weaveworks-ui-components a 0.3.10
    [#2980](https://github.com/weaveworks/scope/pull/2980)
*   Bloquear la versión de componentes de estilo y actualizar los componentes de interfaz de usuario
    [#2976](https://github.com/weaveworks/scope/pull/2976)
*   no incrustar binario de Docker
    [#2977](https://github.com/weaveworks/scope/pull/2977)
*   Bump Embedded Weave Net versión a 2.1.3
    [#2975](https://github.com/weaveworks/scope/pull/2975)
*   no notificar las asignaciones en los índices de referencia
    [#2964](https://github.com/weaveworks/scope/pull/2964)

Mejoras de rendimiento:

*   Teclas de mapa "pasante"
    [#2865](https://github.com/weaveworks/scope/pull/2865)
*   Actualizar informes antes del almacenamiento en caché
    [#2979](https://github.com/weaveworks/scope/pull/2979)
*   informe. Upgrade() agregar implementaciones a los pods como padre
    [#2973](https://github.com/weaveworks/scope/pull/2973)

Cambios relacionados con Weave Cloud:

*   sondeo: use un FQDN absoluto para cloud.weave.works de forma predeterminada
    [#2971](https://github.com/weaveworks/scope/pull/2971)
*   Bump ui-components para incluir el viaje en el tiempo descompuesto
    [#2986](https://github.com/weaveworks/scope/pull/2986)
*   punto de conexión de sonda barato punto de conexión api
    [#2983](https://github.com/weaveworks/scope/pull/2983)

## Versión 1.6.6

Esta es una versión de parche menor.

Nuevas características y mejoras:

*   Actualizar a React 16
    [#2929](https://github.com/weaveworks/scope/pull/2929)
*   Humanizar las duraciones reportadas
    [#2915](https://github.com/weaveworks/scope/pull/2915)
*   Utilice firstSeenConnectAt para el comienzo del período de disponibilidad en Time Travel
    [#2912](https://github.com/weaveworks/scope/pull/2912)
*   Fondo plano en lugar de degradado
    [#2886](https://github.com/weaveworks/scope/pull/2886)

Correcciones de errores y mejoras menores:

*   Corregir informes incorrectos de replicaset DesiredReplicas
    [#2955](https://github.com/weaveworks/scope/pull/2955)
*   filtrar adyacencias de Internet
    [#2954](https://github.com/weaveworks/scope/pull/2954)
*   filtrar pseudonodos no conectados solo una vez, al final
    [#2951](https://github.com/weaveworks/scope/pull/2951)
*   Mostrar correctamente si hay nuevas imágenes o no.
    [#2948](https://github.com/weaveworks/scope/pull/2948)
*   Corregir error de imagen indefinida
    [#2934](https://github.com/weaveworks/scope/pull/2934)
*   Usar marca de tiempo en la URL
    [#2919](https://github.com/weaveworks/scope/pull/2919)
*   Corregir error de texto de estado de imagen incorrecto
    [#2935](https://github.com/weaveworks/scope/pull/2935)
*   No elimine la referencia de los punteros de ECS sin comprobar
    [#2918](https://github.com/weaveworks/scope/pull/2918)
*   MAINTAINER está en desuso, ahora usa LABEL
    [#2916](https://github.com/weaveworks/scope/pull/2916)
*   Intente evitar errores de NaN de zoom
    [#2906](https://github.com/weaveworks/scope/pull/2906)
*   Compruebe si tableContent ref está presente
    [#2907](https://github.com/weaveworks/scope/pull/2907)
*   Usar TimeTravel desde ui-components repo
    [#2903](https://github.com/weaveworks/scope/pull/2903)
*   docker: Cerrar canalización cuando se produce un error en la llamada a la API de Docker
    [#2894](https://github.com/weaveworks/scope/pull/2894)
*   terminal: Corregir error tipográfico caculado
    [#2897](https://github.com/weaveworks/scope/pull/2897)
*   Arreglar golint si/else
    [#2892](https://github.com/weaveworks/scope/pull/2892)
*   Haga de time Travel un componente modular
    [#2888](https://github.com/weaveworks/scope/pull/2888)
*   Utilice una ruta contextual para el terminal emergente.html
    [#2882](https://github.com/weaveworks/scope/pull/2882)
*   Corrige imágenes en el panel de detalles después del servicio -> los recursos cambian
    [#2885](https://github.com/weaveworks/scope/pull/2885)
*   Corrige la funcionalidad "Guardar como svg".
    [#2883](https://github.com/weaveworks/scope/pull/2883)
*   seguimiento: Corregir la colisión en vuelo de dos RELACIONES públicas relacionadas
    [#2867](https://github.com/weaveworks/scope/pull/2867)
*   lint: Arreglar 2 sitios que fallan en una comprobación de golint introducida recientemente
    [#2868](https://github.com/weaveworks/scope/pull/2868)
*   Continuar procesando informes si se produce un error en la facturación
    [#2860](https://github.com/weaveworks/scope/pull/2860)

Documentación:

*   Agregar godoc al archivo README
    [#2891](https://github.com/weaveworks/scope/pull/2891)
*   Año de derechos de autor de la licencia Bump hasta 2017
    [#2873](https://github.com/weaveworks/scope/pull/2873)

Mejoras internas y limpieza:

*   hacer renderizado. ResetCache() restablecer todas las cachés
    [#2949](https://github.com/weaveworks/scope/pull/2949)
*   Decoradores, ¡begone!
    [#2947](https://github.com/weaveworks/scope/pull/2947)
*   Expandir el cálculo del espacio de nombres k8s de la aplicación
    [#2946](https://github.com/weaveworks/scope/pull/2946)
*   Pasar filtros de representación y mapas por valor en lugar de referencia
    [#2944](https://github.com/weaveworks/scope/pull/2944)
*   Quitar dependencias opcionales
    [#2930](https://github.com/weaveworks/scope/pull/2930)
*   eliminar la memoización no utilizada
    [#2924](https://github.com/weaveworks/scope/pull/2924)
*   Simplifique la conversión de cadenas Utsname
    [#2917](https://github.com/weaveworks/scope/pull/2917)
*   vendoring: Actualizar gopacket a la última master
    [#2911](https://github.com/weaveworks/scope/pull/2911)
*   Actualizar dependencias de nodos menores
    [#2900](https://github.com/weaveworks/scope/pull/2900)
*   Actualizar dependencias de eslint
    [#2896](https://github.com/weaveworks/scope/pull/2896)
*   Deshágase de la dependencia react-tooltip
    [#2871](https://github.com/weaveworks/scope/pull/2871)
*   Actualizar ugorji/go/codec
    [#2863](https://github.com/weaveworks/scope/pull/2863)

Mejoras de rendimiento:

*   Hacer que la actualización de informes sea rápida cuando no es una operación
    [#2965](https://github.com/weaveworks/scope/pull/2965)
*   Eliminar el ajuste de estructura LatestMap
    [#2961](https://github.com/weaveworks/scope/pull/2961)
*   Detener la generación de informes de ReplicaSets
    [#2957](https://github.com/weaveworks/scope/pull/2957)
*   optimización: asignar menos memoria en la fusión de LatestMap
    [#2956](https://github.com/weaveworks/scope/pull/2956)
*   Decodificar informes de un búfer de bytes
    [#2386](https://github.com/weaveworks/scope/pull/2386)
*   optimización: más rápido conocidoServiceCache
    [#2952](https://github.com/weaveworks/scope/pull/2952)
*   optimización: hacer que Memoise(Memoise(...)) solo memoise una vez
    [#2950](https://github.com/weaveworks/scope/pull/2950)
*   Reescribir más map-reduce como renderizadores para guardar basura
    [#2938](https://github.com/weaveworks/scope/pull/2938)
*   Optimizaciones de análisis
    [#2937](https://github.com/weaveworks/scope/pull/2937)
*   producir estadísticas como parte de la representación
    [#2926](https://github.com/weaveworks/scope/pull/2926)
*   Optimización: reemplace tres reducciones de mapas con Renderers
    [#2920](https://github.com/weaveworks/scope/pull/2920)
*   Evite la creación de objetos al analizar nombres DNS
    [#2921](https://github.com/weaveworks/scope/pull/2921)
*   Vuelva a implementar LatestMap como un segmento ordenado para un mejor rendimiento
    [#2870](https://github.com/weaveworks/scope/pull/2870)
*   Pequeña reducción de la asignación de memoria
    [#2872](https://github.com/weaveworks/scope/pull/2872)

Cambios relacionados con Weave Cloud:

*   seguimiento: Usar segmento para el seguimiento
    [#2861](https://github.com/weaveworks/scope/pull/2861)
*   seguimiento: Agregar evento Mixpanel en los clics de control
    [#2857](https://github.com/weaveworks/scope/pull/2857)

## Versión 1.6.5

Esta es una versión de parche menor.

Nuevas características y mejoras:

*   Opción Agregar región de clúster de ECS
    [#2854](https://github.com/weaveworks/scope/pull/2854)

Correcciones de errores y mejoras menores:

*   Corregir bordes que desaparecen en el modo gráfico
    [#2851](https://github.com/weaveworks/scope/pull/2851)
*   Capacidad de respuesta de los elementos de encabezado fijos
    [#2839](https://github.com/weaveworks/scope/pull/2839)

Documentación:

*   Actualizar documentos de la AMI de ECS
    [#2846](https://github.com/weaveworks/scope/pull/2846)
*   Instrucciones de composición más precisas
    [#2843](https://github.com/weaveworks/scope/pull/2843)

Mejoras internas y limpieza:

*   Reparar prueba rota por #2854
    [#2856](https://github.com/weaveworks/scope/pull/2856)
*   Corregir la sintaxis de circle.yml
    [#2841](https://github.com/weaveworks/scope/pull/2841)
*   hacer que circleci ui-upload sea incondicional
    [#2837](https://github.com/weaveworks/scope/pull/2837)

Cambios relacionados con Weave Cloud:

*   La conexión de AWS se mantiene viva
    [#2852](https://github.com/weaveworks/scope/pull/2852)

## Versión 1.6.4

Relanzamiento de la versión 1.6.3 para la que hubo algunos problemas para publicar imágenes de Docker.

Documentación:

*   Corregir fraseo en el archivo Léame
    [#2836](https://github.com/weaveworks/scope/pull/2836)

## Versión 1.6.3

Esta es una versión de parche menor.

Nuevas características y mejoras:

*   Hacer que el mensaje URL de Scope sea más preciso
    [#2810](https://github.com/weaveworks/scope/pull/2810)
*   Equilibra la sensibilidad del zoom de la línea de tiempo entre Firefox y Chrome
    [#2788](https://github.com/weaveworks/scope/pull/2788)
*   Ajustar la sensibilidad del zoom de la línea de tiempo en Firefox
    [#2777](https://github.com/weaveworks/scope/pull/2777)
*   Ejecute un shell normal (en lugar de iniciar sesión) en contenedores
    [#2781](https://github.com/weaveworks/scope/pull/2781)
*   Agregar recuento de reinicio de pod al panel de detalles
    [#2761](https://github.com/weaveworks/scope/pull/2761)

Mejoras de rendimiento:

*   Hacer animaciones de gráficos de nodos un poco más rápido
    [#2803](https://github.com/weaveworks/scope/pull/2803)
*   Mejorar el rendimiento de Firefox
    [#2795](https://github.com/weaveworks/scope/pull/2795)
*   sintetizar la red de servicio K8S desde ip de servicio
    [#2779](https://github.com/weaveworks/scope/pull/2779)
    [#2806](https://github.com/weaveworks/scope/pull/2806)

Correcciones de errores y mejoras menores:

*   Reinicie el seguimiento de eBPF en el error
    [#2735](https://github.com/weaveworks/scope/pull/2735)
*   La tabla Fix processes/hosts no aparece
    [#2824](https://github.com/weaveworks/scope/pull/2824)
*   Corregir filtros de consulta agregando espacios de nombres y usando el nombre del contenedor de Docker
    [#2819](https://github.com/weaveworks/scope/pull/2819)
*   Quitar espacios en blanco de listas de conexiones vacías
    [#2811](https://github.com/weaveworks/scope/pull/2811)
*   Corregir la representación de SVG exportado
    [#2794](https://github.com/weaveworks/scope/pull/2794)
*   Sonda k8s: Corregir un pánico (deref de puntero nulo) cuando nunca se ha programado un cronjob
    [#2785](https://github.com/weaveworks/scope/pull/2785)

Documentación:

*   Quitar sangría adicional en la nota
    [#2816](https://github.com/weaveworks/scope/pull/2816)

Mejoras internas y limpieza:

*   Usar Node 8.4 para compilaciones
    [#2830](https://github.com/weaveworks/scope/pull/2830)
*   refactorizar: reducir la duplicación en links_test
    [#2820](https://github.com/weaveworks/scope/pull/2820)
*   Agregar 'RealClean' hacer que el objetivo borre las imágenes del contenedor
    [#2771](https://github.com/weaveworks/scope/pull/2771)
*   deshacerse de los indicadores de tipo de punto final
    [#2772](https://github.com/weaveworks/scope/pull/2772)
*   Cambie el nombre de la capacidad 'report_persistence' por 'historic_reports'
    [#2774](https://github.com/weaveworks/scope/pull/2774)
*   Agregar prueba para el número de pods que no se actualiza
    [#2741](https://github.com/weaveworks/scope/pull/2741)
*   refactorizar: eliminar duplicación
    [#2765](https://github.com/weaveworks/scope/pull/2765)
*   Enviar imágenes de liberación a quay.io
    [#2763](https://github.com/weaveworks/scope/pull/2763)

Cambios relacionados con Weave Cloud:

*   Vincular gráficos scope-ui en los que se puede hacer clic para consultas prometheus
    [#2664](https://github.com/weaveworks/scope/pull/2664)
*   scope/cortex: corregir errores tipográficos en el filtro de consultas
    [#2815](https://github.com/weaveworks/scope/pull/2815)
*   Recorrido en el tiempo: mantenga actualizados los paneles de detalles de nodos activos
    [#2807](https://github.com/weaveworks/scope/pull/2807)
*   Viaje en el tiempo: desmontar en la acción shutdown()
    [#2801](https://github.com/weaveworks/scope/pull/2801)
*   Corrige la entrada de marca de tiempo de viaje en el tiempo que se trunca en OSX
    [#2805](https://github.com/weaveworks/scope/pull/2805)
*   Viaje en el tiempo: elimine el indicador de características y haga que la disponibilidad dependa de la capacidad de informes históricos
    [#2616](https://github.com/weaveworks/scope/pull/2616)
*   Registrar mensajes cuadrados en el nivel 'depurar' en lugar de 'información'
    [#2798](https://github.com/weaveworks/scope/pull/2798)
*   Corregir el desplazamiento vertical de la etiqueta de la línea de tiempo en algunos Chromes
    [#2793](https://github.com/weaveworks/scope/pull/2793)
*   Reduzca la altura de la línea de tiempo y haga que los años se desvanezcan
    [#2778](https://github.com/weaveworks/scope/pull/2778)
*   Siempre tenga en cuenta la marca de tiempo en Detalles del nodo cuando viaje en el tiempo
    [#2775](https://github.com/weaveworks/scope/pull/2775)
*   Ocultar el botón Configurar de la interfaz de usuario del servicio solo en ámbito
    [#2766](https://github.com/weaveworks/scope/pull/2766)
*   Viaje en el tiempo 3.0
    [#2703](https://github.com/weaveworks/scope/pull/2703)

## Versión 1.6.2

Lanzamiento del parche de corrección de errores

*   Sonda k8s: Corregir un pánico (deref de puntero nulo) cuando nunca se ha programado un cronjob
    [#2785](https://github.com/weaveworks/scope/pull/2785)

## Versión 1.6.1

Esta es una relanzamiento de 1.6.0. La compilación oficial para 1.6.0 inadvertidamente
incluía versiones obsoletas de algunos componentes, lo que introducía problemas de seguridad.

## Versión 1.6.0

Resúmenes:

*   Nueva vista de controladores de Kubernetes
*   Agregar conjuntos con estado de Kubernetes y trabajos cron
*   Varias pequeñas mejoras y trabajo de rendimiento

Nuevas características y mejoras:

*   kubernetes: Agregar StatefulSets y CronJobs
    [#2724](https://github.com/weaveworks/scope/pull/2724)
*   Hacer que se pueda hacer clic en los nodos de vista de recursos
    [#2679](https://github.com/weaveworks/scope/pull/2679)
*   Mantener visible la navegación de topología si está seleccionada
    [#2709](https://github.com/weaveworks/scope/pull/2709)
*   Mostrar varios parientes en la vista de cuadrícula de nodos
    [#2648](https://github.com/weaveworks/scope/pull/2648)
*   Quitar filtro de tipo en la vista controladores
    [#2670](https://github.com/weaveworks/scope/pull/2670)
*   Quitar conjuntos de réplicas
    [#2661](https://github.com/weaveworks/scope/pull/2661)
*   Agregar vista combinada de K8S
    [#2552](https://github.com/weaveworks/scope/pull/2552)
*   Recopilar el complemento Weave Net y la información del proxy del informe
    [#2719](https://github.com/weaveworks/scope/pull/2719)

Mejoras de rendimiento:

*   optimización: no copie el flujo de informes innecesariamente
    [#2736](https://github.com/weaveworks/scope/pull/2736)
*   Los nuevos informes completos son más importantes que los informes antiguos y los informes de acceso directo
    [#2743](https://github.com/weaveworks/scope/pull/2743)
*   Aumentar el tamaño predeterminado del búfer Conntrack
    [#2739](https://github.com/weaveworks/scope/pull/2739)
*   Use el nombre del nodo de Kubernetes para filtrar pods si es posible
    [#2556](https://github.com/weaveworks/scope/pull/2556)
*   refactorizar: eliminar Copy() innecesario y muerto
    [#2675](https://github.com/weaveworks/scope/pull/2675)
*   rendimiento: solo color conectado una vez
    [#2635](https://github.com/weaveworks/scope/pull/2635)
*   comprobación rápida de pertenencia a la red
    [#2625](https://github.com/weaveworks/scope/pull/2625)
*   memoize isKnownServices para mejorar el rendimiento
    [#2617](https://github.com/weaveworks/scope/pull/2617)
*   coincidencia más rápida de servicios conocidos
    [#2613](https://github.com/weaveworks/scope/pull/2613)

Correcciones de errores y mejoras menores:

*   k8s: Use 'DaemonSet', 'StatefulSet', etc. en lugar de 'Daemon Set', 'Stateful Set'
    [#2757](https://github.com/weaveworks/scope/pull/2757)
*   maximizar el tiempo de espera de publicación de informes
    [#2756](https://github.com/weaveworks/scope/pull/2756)
*   No retroceda los tiempos de espera al enviar informes
    [#2746](https://github.com/weaveworks/scope/pull/2746)
*   Corregir el número de Pods en el gráfico que no se actualiza (etiqueta secundaria)
    [#2728](https://github.com/weaveworks/scope/pull/2728)
*   defenderse contra los ceros
    [#2734](https://github.com/weaveworks/scope/pull/2734)
*   Corregir la notificación de nueva versión que no se muestra
    [#2720](https://github.com/weaveworks/scope/pull/2720)
*   renderizar: En etiquetas menores, muestre '0 cosas' en lugar de en blanco si no hay cero cosas presentes
    [#2726](https://github.com/weaveworks/scope/pull/2726)
*   Restablecer nodos en frontend cuando se reinicia scope-app
    [#2713](https://github.com/weaveworks/scope/pull/2713)
*   Mantenga visible la navegación del topográfico si se selecciona el subnav
    [#2710](https://github.com/weaveworks/scope/pull/2710)
*   no te pierdas, o no olvides, las conexiones iniciales
    [#2704](https://github.com/weaveworks/scope/pull/2704)
*   bump tcptracer-bpf versión
    [#2705](https://github.com/weaveworks/scope/pull/2705)
*   fix ebpf init race segfault
    [#2695](https://github.com/weaveworks/scope/pull/2695)
*   Hacer que los límites de zoom del diseño del gráfico sean constantes
    [#2678](https://github.com/weaveworks/scope/pull/2678)
*   Última línea de defensa contra nodos superpuestos en el diseño de gráficos
    [#2688](https://github.com/weaveworks/scope/pull/2688)
*   Arreglar `yarn pack` Ignorar el indicador de cli del directorio
    [#2694](https://github.com/weaveworks/scope/pull/2694)
*   Mostrar desbordamiento de tabla solo si el límite excede por 2+
    [#2683](https://github.com/weaveworks/scope/pull/2683)
*   render/pod: Corrige un error tipográfico en Map2Parent donde UnmanagedID siempre se usará para noParentsPseudoID
    [#2685](https://github.com/weaveworks/scope/pull/2685)
*   No mostrar el recuento de contenedores en la lista de imágenes del panel detalle del host
    [#2682](https://github.com/weaveworks/scope/pull/2682)
*   determinación correcta de las imágenes de contenedor de un host
    [#2680](https://github.com/weaveworks/scope/pull/2680)
*   Evita que los pids de 6 dígitos se trunquen en el modo de panel/tabla de detalles
    [#2666](https://github.com/weaveworks/scope/pull/2666)
*   polaridad correcta de las conexiones iniciales
    [#2645](https://github.com/weaveworks/scope/pull/2645)
*   Asegúrese de que las conexiones de /proc/net/tcp{,6} obtengan el PID correcto
    [#2639](https://github.com/weaveworks/scope/pull/2639)
*   Evitar las condiciones de carrera en los dominios almacenados en caché de DNSSnooper
    [#2637](https://github.com/weaveworks/scope/pull/2637)
*   Solucionar problemas con los tipos de sindicatos
    [#2633](https://github.com/weaveworks/scope/pull/2633)
*   Corregir errores tipográficos en site/plugins.md
    [#2624](https://github.com/weaveworks/scope/pull/2624)
*   correcto `nodeSummaryGroupSpec`
    [#2631](https://github.com/weaveworks/scope/pull/2631)
*   Ignorar ipv6
    [#2622](https://github.com/weaveworks/scope/pull/2622)
*   Corregir el error de orden de clasificación de tablas para valores numéricos
    [#2587](https://github.com/weaveworks/scope/pull/2587)
*   Arreglar zoom para `npm start`
    [#2605](https://github.com/weaveworks/scope/pull/2605)
*   corregir el error cuando el DAEMON de docker se ejecuta con el espacio de nombres de usuario habilitado.
    [#2582](https://github.com/weaveworks/scope/pull/2582)
*   No lea archivos tcp6 si TCP versión 6 no es compatible
    [#2604](https://github.com/weaveworks/scope/pull/2604)
*   Credenciales de solo token de Elide en argumentos cli
    [#2593](https://github.com/weaveworks/scope/pull/2593)
*   No filtre los extremos por Procspied/EBPF en los representadores
    [#2652](https://github.com/weaveworks/scope/pull/2652)

Mejoras internas y limpieza:

*   Pasar etiquetas de compilación a pruebas unitarias
    [#2618](https://github.com/weaveworks/scope/pull/2618)
*   Permite omitir la compilación del cliente al hacer make prog / scope
    [#2732](https://github.com/weaveworks/scope/pull/2732)
*   Solo pase WEAVESCOPE_DOCKER_ARGS al inicio real de la sonda/aplicación
    [#2715](https://github.com/weaveworks/scope/pull/2715)
*   Establezca la versión de package.json en 0.0.0
    [#2692](https://github.com/weaveworks/scope/pull/2692)
*   simplificar la unión de conexión
    [#2714](https://github.com/weaveworks/scope/pull/2714)
*   EbpfTracker refactorización / limpieza
    [#2699](https://github.com/weaveworks/scope/pull/2699)
*   Versión de prefijos de hilo con `v` al empacar
    [#2691](https://github.com/weaveworks/scope/pull/2691)
*   no use eBPF en un par de pruebas
    [#2690](https://github.com/weaveworks/scope/pull/2690)
*   Actualizar README/Makefile/package.json para usar hilo
    [#2676](https://github.com/weaveworks/scope/pull/2676)
*   render/pod: Eliminar opciones no utilizadas y código incorrecto
    [#2673](https://github.com/weaveworks/scope/pull/2673)
*   Usar el nuevo cliente k8s go
    [#2659](https://github.com/weaveworks/scope/pull/2659)
*   Actualización github.com/weaveworks/common y dependencias (necesita go1.8)
    [#2570](https://github.com/weaveworks/scope/pull/2570)
*   Publicar instrucciones actualizadas de OpenShift (cierre #2485)
    [#2657](https://github.com/weaveworks/scope/pull/2657)
*   hacer que las pruebas de integración pasen con la última versión de Weave Net (2.0)
    [#2641](https://github.com/weaveworks/scope/pull/2641)
*   Orden de representación mejorado de nodos/aristas en la vista Gráfico
    [#2623](https://github.com/weaveworks/scope/pull/2623)
*   refactorizar: extraer un par de constantes muy utilizadas
    [#2632](https://github.com/weaveworks/scope/pull/2632)
*   Utilice la última versión de go1.8.3
    [#2626](https://github.com/weaveworks/scope/pull/2626)
*   Usar Go 1.8
    [#2621](https://github.com/weaveworks/scope/pull/2621)
*   Use 127.0.0.1 en lugar de localhost, más
    [#2554](https://github.com/weaveworks/scope/pull/2554)
*   Se ha movido la información de nodos/bordes resaltados a los selectores
    [#2584](https://github.com/weaveworks/scope/pull/2584)
*   racionalizar el uso del conjunto de informes
    [#2671](https://github.com/weaveworks/scope/pull/2671)
*   Omitir los extremos con adyacencia >1 en la representación de procesos
    [#2668](https://github.com/weaveworks/scope/pull/2668)
*   Variables env de Honor DOCKER_\* en la sonda y la aplicación
    [#2649](https://github.com/weaveworks/scope/pull/2649)

Cambios relacionados con Weave Cloud:

*   Retroceder al escribir en Dynamo y S3
    [#2723](https://github.com/weaveworks/scope/pull/2723)
*   Rediseño de viajes en el tiempo
    [#2651](https://github.com/weaveworks/scope/pull/2651)
*   Realizar llamadas a la API con marca de tiempo de viaje en el tiempo
    [#2600](https://github.com/weaveworks/scope/pull/2600)

## Versión 1.5.1

Lanzamiento del parche de corrección de errores

Correcciones:

*   las conexiones iniciales tienen una polaridad incorrecta
    [#2644](https://github.com/weaveworks/scope/issues/2644)
*   conexión a proceso muerto asociado a un proceso diferente
    [#2638](https://github.com/weaveworks/scope/pull/2638)

## Versión 1.5.0

Resúmenes:

*   Seguimiento de conexión más preciso y barato con eBPF, que ahora está habilitado de forma predeterminada.
*   Correcciones de errores y mejoras de rendimiento.

Nuevas características y mejoras:

*   Habilitar el seguimiento de eBPF de forma predeterminada
    [#2535](https://github.com/weaveworks/scope/pull/2535)
*   Contraseñas de URL de Elide en argumentos cli
    [#2568](https://github.com/weaveworks/scope/pull/2568)

Mejoras de rendimiento:

*   Suelte el sumador y el puerto desde Endpoint.Mapa más reciente
    [#2581](https://github.com/weaveworks/scope/pull/2581)
*   reducción paralela
    [#2561](https://github.com/weaveworks/scope/pull/2561)
*   No lea todo /proc cuando probe.proc.spy=false
    [#2557](https://github.com/weaveworks/scope/pull/2557)
*   optimizar: no ordenar en NodeSet.ForEach
    [#2548](https://github.com/weaveworks/scope/pull/2548)
*   codificar ps vacíos. Mapas como nulos
    [#2547](https://github.com/weaveworks/scope/pull/2547)

Correcciones:

*   volver a dirigir los clientes de la aplicación cuando cambia la resolución de nombres
    [#2579](https://github.com/weaveworks/scope/pull/2579)
*   tipo correcto para "Generación observada".
    [#2572](https://github.com/weaveworks/scope/pull/2572)
*   Retroceder en las solicitudes de API de kubernetes con errores
    [#2562](https://github.com/weaveworks/scope/pull/2562)
*   Cierre el rastreador eBPF de forma limpia
    [#2541](https://github.com/weaveworks/scope/pull/2541)
*   Simplifique la entrada del rastreador de conexiones y corrija el respaldo de escaneo procfs
    [#2539](https://github.com/weaveworks/scope/pull/2539)
*   Protegerse contra el almacén DaemonSet nulo
    [#2538](https://github.com/weaveworks/scope/pull/2538)

Mejoras internas y limpieza:

*   es6ify server.js e incluir en eslint
    [#2560](https://github.com/weaveworks/scope/pull/2560)
*   Arreglar prog /main_test.go
    [#2567](https://github.com/weaveworks/scope/pull/2567)
*   Corregir dependencias incompletas para `make scope/prog`
    [#2563](https://github.com/weaveworks/scope/pull/2563)
*   Versión de bump package.json a la versión de ámbito actual
    [#2555](https://github.com/weaveworks/scope/pull/2555)
*   simplificar la unión de conexión
    [#2559](https://github.com/weaveworks/scope/pull/2559)
*   Usar ayudantes de mapas
    [#2546](https://github.com/weaveworks/scope/pull/2546)
*   agregar utilidad copyreport
    [#2542](https://github.com/weaveworks/scope/pull/2542)

Cambios relacionados con Weave Cloud:

*   Control de viajes en el tiempo
    [#2524](https://github.com/weaveworks/scope/pull/2524)
*   Agregar capacidades de aplicación al extremo /api
    [#2575](https://github.com/weaveworks/scope/pull/2575)

## Versión 1.4.0

Resúmenes:

*   Nueva vista Docker Swarm
*   Nueva vista DaemonSets de Kubernetes
*   Mejoras en el rendimiento de la sonda
*   Muchas correcciones de errores

Nuevas características y mejoras:

*   Agregar vista Docker Swarm
    [#2444](https://github.com/weaveworks/scope/pull/2444)
    [#2452](https://github.com/weaveworks/scope/pull/2452)
    [#2450](https://github.com/weaveworks/scope/pull/2450)
*   Kuebrnetes: añadir daemonsets
    [#2526](https://github.com/weaveworks/scope/pull/2526)
*   Control de zoom de lienzo
    [#2513](https://github.com/weaveworks/scope/pull/2513)
*   Información de consumo de recursos coherente en la vista de recursos
    [#2499](https://github.com/weaveworks/scope/pull/2499)
*   k8s: mostrar todos los espacios de nombres de forma predeterminada
    [#2522](https://github.com/weaveworks/scope/pull/2522)
*   Ocultar el estado de la imagen del contenedor para pseudonodos
    [#2520](https://github.com/weaveworks/scope/pull/2520)
*   Desglose algunos servicios basados en Azure de "Internet"
    [#2521](https://github.com/weaveworks/scope/pull/2521)
*   Eliminar zoom al hacer doble clic
    [#2457](https://github.com/weaveworks/scope/pull/2457)
*   permitir la desactivación de la publicidad/búsqueda de weaveDNS
    [#2445](https://github.com/weaveworks/scope/pull/2445)

Mejoras de rendimiento:

*   process walker perfs: optimizar readLimits y readStats
    [#2491](https://github.com/weaveworks/scope/pull/2491)
*   proc walker: optimizar el contador de archivos abierto
    [#2456](https://github.com/weaveworks/scope/pull/2456)
*   eliminar las llamadas excesivas a mtime. Ahora()
    [#2486](https://github.com/weaveworks/scope/pull/2486)
*   Msgpack perf: escribir psMap directamente
    [#2466](https://github.com/weaveworks/scope/pull/2466)
*   proc_linux: no exec `getNetNamespacePathSuffix()` en cada paseo
    [#2453](https://github.com/weaveworks/scope/pull/2453)
*   gzip: cambiar el nivel de compresión al valor predeterminado
    [#2437](https://github.com/weaveworks/scope/pull/2437)

Correcciones:

*   Permita que conntrack rastree conexiones de corta duración no NATed
    [#2527](https://github.com/weaveworks/scope/pull/2527)
*   Volver a habilitar los informes de acceso directo de pod
    [#2528](https://github.com/weaveworks/scope/pull/2528)
*   Rastreador de conexión ebpf: correcciones de mapas perf
    [#2507](https://github.com/weaveworks/scope/pull/2507)
*   ebpf: controlar los eventos fdinstall de tcptracer-bpf (también conocido como problema "aceptar antes de kretprobe")
    [#2518](https://github.com/weaveworks/scope/pull/2518)
*   Fijar el posicionamiento de las puntas de flecha
    [#2505](https://github.com/weaveworks/scope/pull/2505)
*   Evitar desreferencias nulas en el cliente ECS
    [#2514](https://github.com/weaveworks/scope/pull/2514)
    [#2515](https://github.com/weaveworks/scope/pull/2515)
*   api_topologies: No coloque filtros de espacio de nombres en contenedores por dns/imagen
    [#2506](https://github.com/weaveworks/scope/pull/2506)
*   Error específico de registro cuando no se admiten implementaciones
    [#2501](https://github.com/weaveworks/scope/pull/2501)
*   Falta la opción de espacio de nombres en el estado de url rompe los filtros
    [#2490](https://github.com/weaveworks/scope/issues/2490)
*   El selector de métricas no muestra la métrica anclada resaltada
    [#2467](https://github.com/weaveworks/scope/issues/2467)
*   Métodos abreviados de teclado de conmutación de modo de vista fijo
    [#2471](https://github.com/weaveworks/scope/pull/2471)
*   no mientas sobre la dirección accesible
    [#2443](https://github.com/weaveworks/scope/pull/2443)
*   Corregir resaltado de nodo para todas las formas
    [#2430](https://github.com/weaveworks/scope/pull/2430)
*   El selector de modo de vista no responde bien al cambio de tamaño
    [#2396](https://github.com/weaveworks/scope/issues/2396)
*   Selector de métricas vacío que aparece como un punto
    [#2425](https://github.com/weaveworks/scope/issues/2425)
*   Borde de nodo de nube demasiado delgado en comparación con otros nodos
    [#2417](https://github.com/weaveworks/scope/issues/2417)
*   Modo de tabla: el origen del panel de detalles no es donde se hace clic
    [#1754](https://github.com/weaveworks/scope/issues/1754)
*   Modo de tabla: la información sobre herramientas para "Internet" carece de una etiqueta menor
    [#1884](https://github.com/weaveworks/scope/issues/1884)
*   Correcciones de carga de viewState desde localStorage en URL
    [#2409](https://github.com/weaveworks/scope/pull/2409)
*   No restablezcas el zoom en el diseño de actualización
    [#2407](https://github.com/weaveworks/scope/pull/2407)
*   Ocultar el panel de ayuda abierto al hacer clic en el icono de la barra de búsqueda
    [#2406](https://github.com/weaveworks/scope/pull/2406)

Documentación:

*   Documentación de la estructura de datos del informe
    [#2025](https://github.com/weaveworks/scope/pull/2025)
*   Agregar documentación de tablas de varias columnas
    [#2516](https://github.com/weaveworks/scope/pull/2516)
*   Instrucciones de instalación de Update k8s
    [#2512](https://github.com/weaveworks/scope/pull/2512)
    [#2519](https://github.com/weaveworks/scope/pull/2519)
*   Actualizar documentos de instalación
    [#2257](https://github.com/weaveworks/scope/pull/2257)
*   Agregar mención de plugin al archivo léame de ámbito
    [#2454](https://github.com/weaveworks/scope/pull/2454)
*   Corregir la desactivación de Ámbito en la AMI de ECS
    [#2435](https://github.com/weaveworks/scope/pull/2435)
*   Agregar documentos de AMI a los documentos principales, instrucciones de token de tejido modificadas en un solo lugar
    [#2307](https://github.com/weaveworks/scope/pull/2307)
    [#2416](https://github.com/weaveworks/scope/pull/2416)
    [#2415](https://github.com/weaveworks/scope/pull/2415)

Mejoras internas y limpieza:

*   Reducir el número de lugares en los que se enumeran explícitamente las topologías
    [#2436](https://github.com/weaveworks/scope/pull/2436)
*   Usar la biblioteca de tipos de prop para silenciar la advertencia de obsolescencia de PropTypes
    [#2498](https://github.com/weaveworks/scope/pull/2498)
*   Actualizar bibliotecas de nodos
    [#2292](https://github.com/weaveworks/scope/pull/2292)
*   Variable de tipo de búsqueda agregada
    [#2493](https://github.com/weaveworks/scope/pull/2493)
*   Agregar vista previa del sitio web a través de Netlify
    [#2480](https://github.com/weaveworks/scope/pull/2480)
*   Espaciado de consisten en encabezados de Markdown
    [#2438](https://github.com/weaveworks/scope/pull/2438)
*   sólo archivos lint en git ls-files, no .git/\*
    [#2477](https://github.com/weaveworks/scope/pull/2477)
*   maestro de publicación en dockerhub (de nuevo)
    [#2449](https://github.com/weaveworks/scope/pull/2449)
*   Script de ámbito: permite que la parte 'usuario' del nombre de la imagen sea dada por DOCKERHUB_USER env var
    [#2447](https://github.com/weaveworks/scope/pull/2447)
*   Crear varios campos anónimos con nombre
    [#2419](https://github.com/weaveworks/scope/pull/2419)
*   proveedor: actualizar gobpf y tcptracer-bpf
    [#2428](https://github.com/weaveworks/scope/pull/2428)
*   actualizaciones y correcciones de extras/marcador
    [#2350](https://github.com/weaveworks/scope/pull/2350)
*   Actualizar tcptracer-bpf y volver a habilitar la prueba 311
    [#2411](https://github.com/weaveworks/scope/pull/2411)
*   Agregar comprobación de opciones antiguas
    [#2405](https://github.com/weaveworks/scope/pull/2405)
*   shfmt: corregir el formato del shell
    [#2533](https://github.com/weaveworks/scope/pull/2533)

Cambios relacionados con Weave Cloud:

*   Cierre el cuerpo de respuesta S3 para evitar fugas
    [#2442](https://github.com/weaveworks/scope/pull/2442)
*   Widget Agregar imágenes de servicio
    [#2487](https://github.com/weaveworks/scope/pull/2487)
*   Agregar métricas de weavenet a la facturación
    [#2504](https://github.com/weaveworks/scope/pull/2504)
*   Calcular las dimensiones de la ventana gráfica desde el div scope-app
    [#2473](https://github.com/weaveworks/scope/pull/2473)
*   Se agregó el seguimiento de mixpanel para algunos eventos básicos
    [#2462](https://github.com/weaveworks/scope/pull/2462)
*   Agregar NodeSeconds al emisor de facturación
    [#2422](https://github.com/weaveworks/scope/pull/2422)

## Versión 1.3.0

Resúmenes:

*   Nueva vista de uso de recursos
*   Nuevas flechas en la vista de gráfico para indicar las direcciones de conexión
*   [Imagen certificada por Docker de Weave Cloud Agent](https://store.docker.com/images/f18f278a-54c1-4f25-b252-6e11112776c5)
*   Seguimiento de conexiones eBPF (habilitado con --probe.ebpf.connections=true)

Nuevas características y mejoras:

*   Vista de uso de recursos
    [#2296](https://github.com/weaveworks/scope/pull/2296)
    [#2390](https://github.com/weaveworks/scope/pull/2390)
*   Flechas de borde
    [#2317](https://github.com/weaveworks/scope/pull/2317)
    [#2342](https://github.com/weaveworks/scope/pull/2342)
*   Agregar seguimiento de conexión eBPF
    [#2135](https://github.com/weaveworks/scope/pull/2135)
    [#2327](https://github.com/weaveworks/scope/pull/2327)
    [#2336](https://github.com/weaveworks/scope/pull/2336)
    [#2366](https://github.com/weaveworks/scope/pull/2366)
*   Ver varios espacios de nombres de Kubernetes a la vez
    [#2404](https://github.com/weaveworks/scope/pull/2404)
*   Excluir contenedores de pausa al representar topologías k8s
    [#2338](https://github.com/weaveworks/scope/pull/2338)
*   Cuando k8s estén presentes, permita el filtrado de contenedores por espacio de nombres
    [#2285](https://github.com/weaveworks/scope/pull/2285)
    [#2348](https://github.com/weaveworks/scope/pull/2348)
    [#2362](https://github.com/weaveworks/scope/pull/2362)
*   Agregar controles de escalado vertical/descendente de ECS Service
    [#2197](https://github.com/weaveworks/scope/pull/2197)
*   Mejorar los informes de errores al invocar el script de tejido
    [#2335](https://github.com/weaveworks/scope/pull/2335)
*   Agregar opciones para ocultar args y env vars
    [#2306](https://github.com/weaveworks/scope/pull/2306)
    [#2311](https://github.com/weaveworks/scope/pull/2311)
    [#2310](https://github.com/weaveworks/scope/pull/2310)
*   Agregar indicador de carga en el cambio de opción de topología
    [#2272](https://github.com/weaveworks/scope/pull/2272)
*   reproducción de informes
    [#2301](https://github.com/weaveworks/scope/pull/2301)
*   Mostrar indicador de carga en los cambios de topología
    [#2232](https://github.com/weaveworks/scope/pull/2232)

Mejoras de rendimiento:

*   Optimizaciones de decodificación de mapas
    [#2364](https://github.com/weaveworks/scope/pull/2364)
*   Elimine LatestMap, para reducir la asignación de memoria
    [#2351](https://github.com/weaveworks/scope/pull/2351)
*   Decodificar a través de un segmento de bytes para memcache y lectura de archivos
    [#2331](https://github.com/weaveworks/scope/pull/2331)
*   quantise informes
    [#2305](https://github.com/weaveworks/scope/pull/2305)
*   Optimizaciones dinámicas de representación de diseño
    [#2221](https://github.com/weaveworks/scope/pull/2221)
    [#2265](https://github.com/weaveworks/scope/pull/2265)

Correcciones:

*   La métrica anclada temporalmente no se muestra en la salida del ratón
    [#2397](https://github.com/weaveworks/scope/issues/2397)
*   La búsqueda no tiene en cuenta los nodos de topologías descargadas
    [#2395](https://github.com/weaveworks/scope/issues/2393)
*   Desbordamiento de altura del panel ayuda en la vista Contenedores
    [#2352](https://github.com/weaveworks/scope/issues/2352)
*   El botón "Guardar lienzo como SVG" se muestra en el modo de tabla
    [#2354](https://github.com/weaveworks/scope/pull/2354)
*   Los procesos sin CMD se muestran sin nombre
    [#2315](https://github.com/weaveworks/scope/issues/2315)
*   Se llama a la animación de pulsación en los nodos de gráficos incluso cuando la consulta de búsqueda no cambia
    [#2255](https://github.com/weaveworks/scope/issues/2255)
*   Faltan nombres de pod
    [#2258](https://github.com/weaveworks/scope/issues/2258)
*   analizar --probe-only según lo previsto
    [#2300](https://github.com/weaveworks/scope/pull/2300)
*   Los estados de zoom de la vista de gráfico se restablecen al cambiar a la vista de tabla
    [#2254](https://github.com/weaveworks/scope/issues/2254)
*   gráfico no renderizado de arriba hacia abajo, a pesar de la falta de ciclos
    [#2267](https://github.com/weaveworks/scope/issues/2267)
*   Ocultar filtro no contenido en la vista DNS no ocultar sin contenido
    [#2170](https://github.com/weaveworks/scope/issues/2170)

Documentación:

*   Mejoras en la documentación
    [#2252](https://github.com/weaveworks/scope/pull/2252)
*   Se ha eliminado el texto de combinación perdido y se ha hecho coherente la terminología
    [#2289](https://github.com/weaveworks/scope/pull/2289)

Mejoras internas y limpieza:

*   prueba de integración: deshabilitar la prueba escamosa 311
    [#2380](https://github.com/weaveworks/scope/pull/2380)
*   Agregar trabajo para desencadenar la compilación de la interfaz de usuario de servicio
    [#2376](https://github.com/weaveworks/scope/pull/2376)
*   Usar el administrador de paquetes de hilo
    [#2368](https://github.com/weaveworks/scope/pull/2368)
*   pruebas de integración: enumerar contenedores para depurar
    [#2346](https://github.com/weaveworks/scope/pull/2346)
*   ámbito: use los mismos args de Docker para la ejecución en seco temprana
    [#2326](https://github.com/weaveworks/scope/pull/2326)
    [#2358](https://github.com/weaveworks/scope/pull/2358)
*   Versión Bump react
    [#2339](https://github.com/weaveworks/scope/pull/2339)
*   integración: deshabilitar pruebas con Internet Edge
    [#2314](https://github.com/weaveworks/scope/pull/2314)
*   Pruebas de integración seguras
    [#2312](https://github.com/weaveworks/scope/pull/2312)
*   integración: reinicie el demonio de Docker después de cada prueba
    [#2298](https://github.com/weaveworks/scope/pull/2298)
*   Se ha cambiado el trabajo ui-build-pkg para usar un contenedor docker
    [#2281](https://github.com/weaveworks/scope/pull/2281)
*   pruebas de integración: corregir scripts
    [#2225](https://github.com/weaveworks/scope/pull/2225)
*   circle.yml: Corrija el paso de carga de la interfaz de usuario para que no se cree dos veces
    [#2266](https://github.com/weaveworks/scope/pull/2266)

Cambios relacionados con Weave Cloud:

*   Crear imagen de agente en la nube
    [#2284](https://github.com/weaveworks/scope/pull/2284)
    [#2277](https://github.com/weaveworks/scope/pull/2277)
    [#2278](https://github.com/weaveworks/scope/pull/2278)
*   Los segundos de contenedor no deben ser nanosegundos de contenedor
    [#2372](https://github.com/weaveworks/scope/pull/2372)
*   Borrar el sondeo del cliente y el estado de los nodos al desmontar
    [#2361](https://github.com/weaveworks/scope/pull/2361)
*   Emisor de facturación fluida
    [#2359](https://github.com/weaveworks/scope/pull/2359)
*   Etiqueta métrica de dynamoDB correcta
    [#2344](https://github.com/weaveworks/scope/pull/2344)
*   Agregar lógica para desactivar las solicitudes de red cuando Scope se desmonta
    [#2290](https://github.com/weaveworks/scope/pull/2290)
    [#2340](https://github.com/weaveworks/scope/pull/2340)
*   Cargar hoja de estilos de contraste
    [#2256](https://github.com/weaveworks/scope/pull/2256)
*   Consolide las solicitudes de API en un solo ayudante; Se ha añadido el encabezado CSRF
    [#2260](https://github.com/weaveworks/scope/pull/2260)
*   Agregar lógica para eliminar el estado no transferible al cambiar de instancias de Cloud
    [#2237](https://github.com/weaveworks/scope/pull/2237)

## Versión 1.2.1

Esta es una versión de parche menor.

Documentación

*   Nueva captura de pantalla del token en la nube cargada
    [#2248](https://github.com/weaveworks/scope/pull/2248)
*   Descripción actualizada de la nube
    [#2249](https://github.com/weaveworks/scope/pull/2249)

Correcciones de errores

*   Corregir el menú de ayuda que no se abre desde la sugerencia 'buscar'
    [#2230](https://github.com/weaveworks/scope/pull/2230)
*   Refactorizar el código de generación de URL de la API
    [#2202](https://github.com/weaveworks/scope/pull/2202)

Mejoras

*   Reintroducir los indicadores de punto de control de sondeo para la versión del kernel y el sistema operativo
    [#2224](https://github.com/weaveworks/scope/pull/2224)
*   Xterm.js actualizado a 2.2.3
    [#2126](https://github.com/weaveworks/scope/pull/2126)
*   Permitir semilla aleatoria en el marcador
    [#2206](https://github.com/weaveworks/scope/pull/2206)
*   Cambiar el nombre de los identificadores de nodo de servicio ECS para que sean cluster;serviceName
    [#2186](https://github.com/weaveworks/scope/pull/2186)

## Versión 1.2.0

Resúmenes:

*   Mejoras de rendimiento (tanto en la interfaz de usuario como en las sondas).
*   Scope ahora requiere la versión de Docker >= 1.10.

Nuevas características y mejoras:

*   ECS: el panel de detalles del servicio debe enumerar sus tareas
    [#2041](https://github.com/weaveworks/scope/issues/2041)
*   Priorizar las topologías ecs en la carga inicial si está disponible
    [#2105](https://github.com/weaveworks/scope/pull/2105)
*   Agregar icono de estado de control al encabezado Terminal
    [#2087](https://github.com/weaveworks/scope/pull/2087)
*   Mejoras en el script de inicio de ámbito
    [#2077](https://github.com/weaveworks/scope/pull/2077)
    [#2093](https://github.com/weaveworks/scope/pull/2093)
*   Mantener el enfoque en las filas de tabla de nodos flotantes
    [#2115](https://github.com/weaveworks/scope/pull/2115)
*   Agregar control para restablecer el estado de la vista local
    [#2080](https://github.com/weaveworks/scope/pull/2080)
*   Compruebe que los eventos conntrack están habilitados en el kernel
    [#2112](https://github.com/weaveworks/scope/pull/2112)
*   Hardcode 127.0.0.1 como IP de bucle invertido para el destino predeterminado
    [#2103](https://github.com/weaveworks/scope/pull/2103)
*   prog/main: use flags.app.port para el destino predeterminado
    [#2096](https://github.com/weaveworks/scope/pull/2096)

Mejoras de rendimiento:

*   Optimizaciones de diseño de gráficos
    [#2128](https://github.com/weaveworks/scope/pull/2128)
    [#2179](https://github.com/weaveworks/scope/pull/2179)
    [#2180](https://github.com/weaveworks/scope/pull/2180)
    [#2210](https://github.com/weaveworks/scope/pull/2210)
*   Deshabilitar XML en el análisis de conntrack
    [#2095](https://github.com/weaveworks/scope/pull/2095)
    [#2118](https://github.com/weaveworks/scope/pull/2118)

Correcciones:

*   Reportero de ECS limitado por la API de AWS
    [#2050](https://github.com/weaveworks/scope/issues/2050)
*   Conexiones ya cerradas que aparecen en la pestaña contenedores
    [#2181](https://github.com/weaveworks/scope/issues/2181)
*   Detalles del nodo spinner Chrome display bug fix
    [#2177](https://github.com/weaveworks/scope/pull/2177)
*   corregir el error cuando el demonio de Docker se ejecuta con el espacio de nombres de usuario habilitado.
    [#2161](https://github.com/weaveworks/scope/pull/2161)
    [#2176](https://github.com/weaveworks/scope/pull/2176)
*   DNSSnooper: Admite Dot1Q y limita los errores de decodificación
    [#2155](https://github.com/weaveworks/scope/issues/2155)
*   El modo de contraste no funciona
    [#2165](https://github.com/weaveworks/scope/issues/2165)
    [#2138](https://github.com/weaveworks/scope/issues/2138)
*   El ámbito no crea nodos especiales dentro de la misma VPC
    [#2163](https://github.com/weaveworks/scope/issues/2163)
*   La vista predeterminada no selecciona 'Sólo contenedores de aplicaciones'
    [#2120](https://github.com/weaveworks/scope/issues/2120)
*   ECS: Falta el vínculo con la tarea en el panel de detalles del contenedor
    [#2040](https://github.com/weaveworks/scope/issues/2040)
*   El reportero de kubernetes está roto en katacoda
    [#2049](https://github.com/weaveworks/scope/pull/2049)
*   El procspy de la sonda no informa de las conexiones de larga duración semidúplex de Netcat
    [#1972](https://github.com/weaveworks/scope/issues/1972)
*   El componente Sparkline produce errores cuando se apaga un contenedor
    [#2072](https://github.com/weaveworks/scope/pull/2072)
*   Los botones de gráfico/tabla no cambian de tamaño
    [#2056](https://github.com/weaveworks/scope/issues/2056)
*   Error JS en bordes con muchos waypoints
    [#1187](https://github.com/weaveworks/scope/issues/1187)
*   Corregir dos errores causados por la transición a D3 v4
    [#2048](https://github.com/weaveworks/scope/pull/2048)
*   Los estilos de terminal emergentes no se alinean del todo con los estilos de terminal dentro del alcance
    [#2209](https://github.com/weaveworks/scope/issues/2209)
*   Los radios de forma de esquina redondeada no se alinean del todo
    [#2212](https://github.com/weaveworks/scope/issues/2212)

Documentación:

*   Corregir argumentos de ámbito en los documentos de instalación de Docker Compose
    [#2143](https://github.com/weaveworks/scope/pull/2143)
*   Documentar cómo ejecutar pruebas en el sitio web
    [#2131](https://github.com/weaveworks/scope/pull/2131)
*   Siga las redirecciones en curl al obtener recursos k8s
    [#2067](https://github.com/weaveworks/scope/pull/2067)

Mejoras internas y limpieza:

*   Incrustar y requerir Docker >= 1.10
    [#2190](https://github.com/weaveworks/scope/pull/2190)
*   no intente hacer que "limpiar" funcione en pagos antiguos
    [#2189](https://github.com/weaveworks/scope/pull/2189)
*   Corregir errores de linter
    [#2068](https://github.com/weaveworks/scope/pull/2068)
    [#2166](https://github.com/weaveworks/scope/pull/2166)
*   Solucionar problemas de propiedad con cliente/build-external
    [#2153](https://github.com/weaveworks/scope/pull/2153)
*   Permitir que la interfaz de usuario de ámbito se instale como un módulo de nodo
    [#2144](https://github.com/weaveworks/scope/pull/2144)
    [#2159](https://github.com/weaveworks/scope/pull/2159)
*   Actualizar la imagen base del contenedor a alpine:3.5
    [#2158](https://github.com/weaveworks/scope/pull/2158)
*   Usa Sass en lugar de Less
    [#2141](https://github.com/weaveworks/scope/pull/2141)
*   sonda: sonda de refactorizaciónMain
    [#2148](https://github.com/weaveworks/scope/pull/2148)
*   Actualización a go1.7.4
    [#2147](https://github.com/weaveworks/scope/pull/2147)
*   Pruebas de integración de subárboles y correcciones de herramientas bump
    [#2136](https://github.com/weaveworks/scope/pull/2136)
*   Agregar compatibilidad con tablas genéricas de varias columnas
    [#2109](https://github.com/weaveworks/scope/pull/2109)
*   extras/marcador: mover dialer.go al subdirectorio
    [#2108](https://github.com/weaveworks/scope/pull/2108)
*   Reenviar la versión del sistema operativo/kernel al punto de control
    [#2101](https://github.com/weaveworks/scope/pull/2101)
*   Arreglar forzar el empuje al maestro
    [#2094](https://github.com/weaveworks/scope/pull/2094)
*   Eslint actualizado y eslint-config-airbnb
    [#2058](https://github.com/weaveworks/scope/pull/2058)
    [#2084](https://github.com/weaveworks/scope/pull/2084)
    [#2089](https://github.com/weaveworks/scope/pull/2089)
*   ecs reporter: Corregir algunas líneas de registro que pasaban \*string en lugar de string
    [#2060](https://github.com/weaveworks/scope/pull/2060)
*   Agregar indicador para el registro de encabezados
    [#2086](https://github.com/weaveworks/scope/pull/2086)
*   Añadir extras/marcador
    [#2082](https://github.com/weaveworks/scope/pull/2082)
*   Eliminar wcloud
    [#2081](https://github.com/weaveworks/scope/pull/2081)
*   Agregar linting de cliente a la configuración de CI
    [#2076](https://github.com/weaveworks/scope/pull/2076)
*   Importación explícita de funciones de utilidad lodash
    [#2053](https://github.com/weaveworks/scope/pull/2053)
*   procspy: utilice un lector para copiar el búfer del lector en segundo plano
    [#2020](https://github.com/weaveworks/scope/pull/2020)
*   Usar repositorio "común" recién creado
    [#2061](https://github.com/weaveworks/scope/pull/2061)
*   Corregir todas las versiones de la biblioteca npm
    [#2057](https://github.com/weaveworks/scope/pull/2057)
*   linter: corregir puntuación y mayúsculas
    [#2021](https://github.com/weaveworks/scope/pull/2021)
*   Usando `webpack-dev-middleware` En lugar de `webpack-dev-server` directamente
    [#2034](https://github.com/weaveworks/scope/pull/2034)
*   Crear `latest_release` Etiqueta de imagen de Docker durante el proceso de lanzamiento
    [#2216](https://github.com/weaveworks/scope/issues/2216)

Cambios relacionados con Weave Cloud:

*   Implementar en el muelle al fusionar con el maestro
    [#2134](https://github.com/weaveworks/scope/pull/2134)
*   Se ha eliminado la barra diagonal inicial de la solicitud de API getAllNodes()
    [#2124](https://github.com/weaveworks/scope/pull/2124)
*   Apretones de manos de websocket correctamente instrumentados
    [#2074](https://github.com/weaveworks/scope/pull/2074)

## Versión 1.1.0

Resúmenes:

*   Nueva vista ECS que permite visualizar sus tareas y servicios en EC2 Container Service de Amazon.
*   Los filtros de contenedor personalizados basados en etiquetas se pueden definir a través de `--app.container-label-filter`

Nuevas características y mejoras:

*   Agregar vistas de ECS
    [#2026](https://github.com/weaveworks/scope/pull/2026)
*   Agregar filtros personalizados basados en etiquetas en la vista de contenedor
    [#1895](https://github.com/weaveworks/scope/pull/1895)
*   Mejorar la información sobre herramientas de errores de plugins
    [#2022](https://github.com/weaveworks/scope/pull/2022)
*   Agregar heurística anti-danza (y banderas de características)
    [#1993](https://github.com/weaveworks/scope/pull/1993)
*   Modo de tabla: ordenar ips numéricamente
    [#2007](https://github.com/weaveworks/scope/pull/2007)
*   Aumentar el contraste de texto en blanco y negro en el modo de contraste
    [#2006](https://github.com/weaveworks/scope/pull/2006)
*   Mejore la usabilidad del botón view-node-in-topo
    [#1926](https://github.com/weaveworks/scope/pull/1926)
*   Ocultar topología de tejido si está vacía
    [#2035](https://github.com/weaveworks/scope/pull/2035)

Mejoras de rendimiento:

*   Agregar comprobación de complejidad de gráficos en la carga de la página
    [#1994](https://github.com/weaveworks/scope/pull/1994)

Correcciones:

*   tapón goroutine fuga en el control
    [#2003](https://github.com/weaveworks/scope/pull/2003)
*   Corregir detalles panel que no se cierra en el lienzo haga clic en
    [#1998](https://github.com/weaveworks/scope/pull/1998)
*   Se necesita una ruta pública vacía para las rutas relativas de ámbito
    [#2043](https://github.com/weaveworks/scope/pull/2043)

Documentación:

*   Usar un nombre de servicio independiente intuitivo en la redacción
    [#2019](https://github.com/weaveworks/scope/pull/2019)
*   Corregir el comando kubectl port-forward para acceder a la aplicación scope localmente
    [#2010](https://github.com/weaveworks/scope/pull/2010)
*   Actualizar la documentación de los plugins del sitio web
    [#2008](https://github.com/weaveworks/scope/pull/2008)

Mejoras internas y limpieza:

*   Archivos de configuración combinados externos y prod webpack
    [#2014](https://github.com/weaveworks/scope/pull/2014)
*   Actualizar package.json
    [#2017](https://github.com/weaveworks/scope/pull/2017)
*   Mover plugins a la nueva organización
    [#1906](https://github.com/weaveworks/scope/pull/1906)
*   Cambiar la configuración local de webpack para usar mapas de origen
    [#2011](https://github.com/weaveworks/scope/pull/2011)
*   MiddleWare/ErrorHandler: Implemente Hijacker para que funcione con el proxy WS
    [#1971](https://github.com/weaveworks/scope/pull/1971)
*   Corrección de la prueba dependiente del tiempo (dejar de probar la biblioteca de cliente de Docker)
    [#2005](https://github.com/weaveworks/scope/pull/2005)
*   Dé tiempo a que los colectores de retroceso de prueba de superposición terminen
    [#1995](https://github.com/weaveworks/scope/pull/1995)
*   Actualizar D3 a la versión 4.4.0
    [#2028](https://github.com/weaveworks/scope/pull/2028)

Cambios relacionados con Weave Cloud:

*   Agregar compatibilidad con OpenTracing a TimeRequestHistogram
    [#2023](https://github.com/weaveworks/scope/pull/2023)

## Versión 1.0.0

Resúmenes:

*   Nueva vista de red weave que permite visualizar y solucionar problemas de su red Weave.
*   Nuevos nodos para servicios conocidos. El nodo de Internet ahora se divide en nodos individuales para servicios en la nube conocidos.
*   Terminales mejorados, con un cambio de tamaño adecuado, bloqueo de desplazamiento y mejores imágenes.
*   Interfaz de usuario refinada con información de conexión particularmente mejorada.
*   Muchos bichos aplastados.

Nuevas características y mejoras:

*   Nueva vista de Weave Net
    [#1182](https://github.com/weaveworks/scope/pull/1182)
    [#1973](https://github.com/weaveworks/scope/pull/1973)
    [#1981](https://github.com/weaveworks/scope/pull/1981)
*   Mostrar servicios conocidos
    [#1863](https://github.com/weaveworks/scope/pull/1863)
    [#1881](https://github.com/weaveworks/scope/pull/1881)
    [#1887](https://github.com/weaveworks/scope/pull/1887)
    [#1897](https://github.com/weaveworks/scope/pull/1897)
*   Mejoras en el terminal
    *   Cambiar el tamaño de los TTY
        [#1966](https://github.com/weaveworks/scope/pull/1966)
        [#1979](https://github.com/weaveworks/scope/pull/1979)
        [#1976](https://github.com/weaveworks/scope/pull/1976)
    *   Habilitar el bloqueo de desplazamiento en el terminal
        [#1932](https://github.com/weaveworks/scope/pull/1932)
    *   Agrega información sobre herramientas al botón emergente de terminal
        [#1790](https://github.com/weaveworks/scope/pull/1790)
    *   Clarificar terminal es una ventana secundaria del panel de detalles.
        [#1903](https://github.com/weaveworks/scope/pull/1903)
    *   Usar shells de inicio de sesión en terminales
        [#1821](https://github.com/weaveworks/scope/pull/1821)
*   Mejoras varias en la interfaz de usuario
    *   Mostrar más detalles de las conexiones a Internet de un nodo
        [#1875](https://github.com/weaveworks/scope/pull/1875)
    *   Cerrar cuadro de diálogo de ayuda cuando se hace clic en el lienzo
        [#1960](https://github.com/weaveworks/scope/pull/1960)
    *   Mejorar el formato de 'fecha' de la tabla de metadatos
        [#1927](https://github.com/weaveworks/scope/pull/1927)
    *   Agregar una nueva sección de búsqueda a la ventana emergente de ayuda
        [#1919](https://github.com/weaveworks/scope/pull/1919)
    *   Agregar label_minor a la información sobre herramientas en la tabla conexiones
        [#1912](https://github.com/weaveworks/scope/pull/1912)
    *   Agregar compatibilidad con localstorage para guardar el estado de la vista
        [#1853](https://github.com/weaveworks/scope/pull/1853)
    *   Convierte los servicios en la topología inicial, si está disponible
        [#1823](https://github.com/weaveworks/scope/pull/1823)
    *   Agregar tabla de información de imagen al panel de detalles del contenedor
        [#1942](https://github.com/weaveworks/scope/pull/1942)
*   Permita que el usuario especifique direcciones URL en la línea de comandos y úselas para permitir tokens por destino.
    [#1901](https://github.com/weaveworks/scope/pull/1901)
*   Aplicar filtros de la vista actual al panel de detalles
    [#1904](https://github.com/weaveworks/scope/pull/1904)
*   Aumente la precisión de la marca de tiempo
    [#1933](https://github.com/weaveworks/scope/pull/1933)
*   Agregue el extremo de métricas de prometheus a las sondas.
    [#1915](https://github.com/weaveworks/scope/pull/1915)
*   Permitir a los usuarios especificar el tamaño del búfer conntrack.
    [#1896](https://github.com/weaveworks/scope/pull/1896)
*   Plugins: Añadir soporte para controles basados en tablas
    [#1818](https://github.com/weaveworks/scope/pull/1818)

Mejoras de rendimiento:

*   Hacer que smartMerger.Combine informes de combinación en paralelo
    [#1827](https://github.com/weaveworks/scope/pull/1827)

Correcciones:

*   Fuga de Goroutine en la aplicación scope
    [#1916](https://github.com/weaveworks/scope/issues/1916)
    [#1920](https://github.com/weaveworks/scope/pull/1920)
*   El uso de CPU no es preciso en los hosts
    [#1664](https://github.com/weaveworks/scope/issues/664)
*   Ciertas cadenas de consulta contendrían un && en lugar de &
    [#1953](https://github.com/weaveworks/scope/pull/1953)
*   Las métricas en el lienzo se atascan
    [#1829](https://github.com/weaveworks/scope/issues/1829)
*   conntrack no se usa aunque esté funcionando
    [#1826](https://github.com/weaveworks/scope/issues/1826)
*   Los recuentos de pods y las listas de paneles de detalles no respetan el espacio de nombres
    [#1824](https://github.com/weaveworks/scope/issues/1824)
*   Descartar conexiones de corta duración hacia/desde Pods en la red host
    [#1944](https://github.com/weaveworks/scope/pull/1944)
*   sondeo: la recopilación de estadísticas se puede iniciar dos veces
    [#1799](https://github.com/weaveworks/scope/issues/1799)
*   Error visual donde aparece el lapso vacío
    [#1945](https://github.com/weaveworks/scope/issues/1945)
*   Los recuentos de conexiones entrantes a Internet son demasiado detallados
    [#1867](https://github.com/weaveworks/scope/issues/1867)
*   Dirección IP truncada en la lista de conexiones del panel detalles del nodo de Internet
    [#1862](https://github.com/weaveworks/scope/issues/1862)
*   Número incorrecto de conexiones que se muestran en los nodos de Internet
    [#1495](https://github.com/weaveworks/scope/issues/1495)
*   Detalles Los recuentos de conexiones del panel son demasiado altos
    [#1842](https://github.com/weaveworks/scope/issues/1842)
*   Las conexiones entrantes a Internet se resuelven incorrectamente
    [#1847](https://github.com/weaveworks/scope/issues/1847)
*   El ámbito se bloquea después de la recarga del explorador si la topología actual desaparece
    [#1880](https://github.com/weaveworks/scope/issues/1880)
*   Nombres de nodo en la lista de conexiones truncados innecesariamente
    [#1882](https://github.com/weaveworks/scope/issues/1882)
*   Los valores numéricos de las tablas del panel detalles deben estar alineados a la derecha
    [#1794](https://github.com/weaveworks/scope/issues/1794)
*   La línea de estado del plugin está rota
    [#1825](https://github.com/weaveworks/scope/issues/1825)
*   Modo de tabla: las columnas no métricas se ordenan alfabéticamente al revés
    [#1802](https://github.com/weaveworks/scope/issues/1802)
*   Corregir el escape del argumento en Scope
    [#1950](https://github.com/weaveworks/scope/pull/1950)
*   El panel detalles de la imagen muestra el nombre de la imagen truncada en lugar del ID
    [#1835](https://github.com/weaveworks/scope/issues/1835)
*   Información sobre herramientas truncada
    [#1139](https://github.com/weaveworks/scope/issues/1139)
*   Altura incorrecta de la ventana del terminal en Safari
    [#1986](https://github.com/weaveworks/scope/issues/1986)

Documentación:

*   Simplifique las instrucciones de k8s
    [#1886](https://github.com/weaveworks/scope/pull/1886)
*   Mejore la documentación de instalación
    [#1838](https://github.com/weaveworks/scope/pull/1838)
*   Actualizar la versión de ámbito en la documentación
    [#1859](https://github.com/weaveworks/scope/pull/1859)

Mejoras internas y limpieza:

*   La aplicación se apaga correctamente, lo que permite que las solicitudes http activas finalicen con el tiempo de espera
    [#1839](https://github.com/weaveworks/scope/pull/1839)
*   middleware/errorhandler: Corrige un error que significa que nunca funciona
    [#1958](https://github.com/weaveworks/scope/pull/1958)
*   middleware: Agregar un middleware ErrorHandler utilizado para servir a un controlador alternativo en un determinado código de error
    [#1954](https://github.com/weaveworks/scope/pull/1954)
*   Actualizar deps de cliente para usar Node v6.9.0
    [#1959](https://github.com/weaveworks/scope/pull/1959)
*   Cambie la biblioteca de términos.js a xterm.js
    [#1948](https://github.com/weaveworks/scope/pull/1948)
*   Corregir errores de linter en campos sin clave
    [#1922](https://github.com/weaveworks/scope/pull/1922)
*   Corrija el error linter para la cadena en contexto. WithValue
    [#1921](https://github.com/weaveworks/scope/pull/1921)
*   Subárbol de herramientas de actualización
    [#1937](https://github.com/weaveworks/scope/pull/1937)
*   Corrija circle.yml para implementar realmente los cambios en la interfaz de usuario externa
    [#1910](https://github.com/weaveworks/scope/pull/1910)
*   Extender el middleware de registro para que, opcionalmente, solo registre las solicitudes HTTP fallidas
    [#1909](https://github.com/weaveworks/scope/pull/1909)
*   Agregar opción al ámbito para que se sirva contenido estático desde S3 en su lugar
    [#1908](https://github.com/weaveworks/scope/pull/1908)
*   Actualizar a go1.7
    [#1797](https://github.com/weaveworks/scope/pull/1797)
*   circleci: insertar la imagen del complemento de control de tráfico en docker hub
    [#1858](https://github.com/weaveworks/scope/pull/1858)
*   refactorizar: pluralización de extractos
    [#1855](https://github.com/weaveworks/scope/pull/1855)
*   utilice Client.Stats de go-dockerclient
    [#1833](https://github.com/weaveworks/scope/pull/1833)
*   Imprimir registros para depurar la prueba de integración de apagado
    [#1888](https://github.com/weaveworks/scope/pull/1888)
*   Permitir un RouteMatcher nulo en la instrumentación
    [#1852](https://github.com/weaveworks/scope/pull/1852)

Cambios relacionados con Weave Cloud:

*   No volver a codificar informes en el recopilador
    [#1819](https://github.com/weaveworks/scope/pull/1819)

## Versión 0.17.1

Esta es una versión de parche menor.

Nuevas características y mejoras:

*   Extender los indicadores de cliente de kubernetes para que coincidan con kubectl
    [#1813](https://github.com/weaveworks/scope/pull/1813)

Correcciones:

*   Corregir la superposición de etiquetas de nodo
    [#1812](https://github.com/weaveworks/scope/pull/1812)
*   Arreglar `scope stop` en Docker para Mac
    [#1811](https://github.com/weaveworks/scope/pull/1811)

## Versión 0.17.0

Resúmenes:

*   Nuevo modo de tabla como alternativa a la vista gráfica clásica de Scope. Proporciona
    mayor densidad de información, resultando particularmente útil cuando hay muchos
    nodos en la vista de gráfico.
*   Mejoras considerables en el rendimiento: la eficiencia de la CPU de la aplicación Scope tiene
    aumentó en más del 50% y las sondas Scope en más del 25%.

Nuevas características y mejoras:

*   Modo tabla
    [#1673](https://github.com/weaveworks/scope/pull/1673)
    [#1747](https://github.com/weaveworks/scope/pull/1747)
    [#1753](https://github.com/weaveworks/scope/pull/1753)
    [#1774](https://github.com/weaveworks/scope/pull/1774)
    [#1775](https://github.com/weaveworks/scope/pull/1775)
    [#1784](https://github.com/weaveworks/scope/pull/1784)
*   Indicador de carga
    [#1485](https://github.com/weaveworks/scope/pull/1485)
*   No mostrar el logotipo del tejido cuando se ejecuta en un marco
    [#1734](https://github.com/weaveworks/scope/pull/1734)
*   Reducir el espacio horizontal entre nodos en las vistas de topología
    [#1693](https://github.com/weaveworks/scope/pull/1693)
*   Elide service-token al registrar argumentos de línea de comandos
    [#1782](https://github.com/weaveworks/scope/pull/1782)
*   No te quejes al detener Scope si no se estaba ejecutando
    [#1783](https://github.com/weaveworks/scope/pull/1783)
*   Silenciar el cierre anormal del websocket
    [#1768](https://github.com/weaveworks/scope/pull/1768)
*   Elimine el ruido del registro de estadísticas de los contenedores detenidos
    [#1687](https://github.com/weaveworks/scope/pull/1687)
    [#1798](https://github.com/weaveworks/scope/pull/1798)
*   Ocultar no contenido/no administrado de forma predeterminada
    [#1694](https://github.com/weaveworks/scope/pull/1694)

Mejoras de rendimiento:

*   Eliminar y optimizar más Copy()s
    [#1739](https://github.com/weaveworks/scope/pull/1739)
*   Usar sectores en lugar de listas vinculadas para Métrica
    [#1732](https://github.com/weaveworks/scope/pull/1732)
*   No copiar() en Merge()
    [#1728](https://github.com/weaveworks/scope/pull/1728)
*   Mejore el rendimiento de los mapas inmutables
    [#1720](https://github.com/weaveworks/scope/pull/1720)
*   Codificador personalizado para los mapas más recientes
    [#1709](https://github.com/weaveworks/scope/pull/1709)

Correcciones:

*   Conexiones dentro de un contenedor que se muestran como si fueran entre contenedores
    [#1733](https://github.com/weaveworks/scope/issues/1733)
*   Las sondas filtran dos goroutines al cerrar la ventana de conexión/exec
    [#1767](https://github.com/weaveworks/scope/issues/1767)
*   Escale las etiquetas de nodo con el tamaño del nodo.
    [#1773](https://github.com/weaveworks/scope/pull/1773)
*   Los contenedores infra de Kubernetes parecen resurgir en la última versión 1.3
    [#1750](https://github.com/weaveworks/scope/issues/1750)
*   El icono de búsqueda está encima del texto
    [#1715](https://github.com/weaveworks/scope/issues/1715)
*   Resaltar es impredecible
    [#1756](https://github.com/weaveworks/scope/pull/1520)
*   El panel Detalles trunca el puerto a cuatro dígitos
    [#1711](https://github.com/weaveworks/scope/issues/1711)
*   Contenedores detenidos que no se muestran con sus nombres
    [#1691](https://github.com/weaveworks/scope/issues/1691)
*   Los terminales no admiten caracteres de citas de distribuciones de teclado internacional
    [#1403](https://github.com/weaveworks/scope/issues/1403)

Mejoras internas y limpieza:

*   Script del iniciador: Corregir espacios en blanco incoherentes
    [#1776](https://github.com/weaveworks/scope/pull/1776)
*   Correcciones de pelusa
    [#1751](https://github.com/weaveworks/scope/pull/1751)
*   Agregar registro de la consola del navegador para websocket para renderizar tiempos
    [#1742](https://github.com/weaveworks/scope/pull/1742)
*   circle.yml: implementar maestro con cuentas de concentrador que no sean ascendentes
    [#1655](https://github.com/weaveworks/scope/pull/1655)
    [#1710](https://github.com/weaveworks/scope/pull/1710)
*   Eliminar código de instrumentación no utilizado
    [#1722](https://github.com/weaveworks/scope/pull/1722)
*   Actualizar la versión de las herramientas de compilación
    [#1685](https://github.com/weaveworks/scope/pull/1685)
*   Agregar indicador para la generación de perfiles de bloques
    [#1681](https://github.com/weaveworks/scope/pull/1681)

Cambios relacionados con Weave Cloud:

*   Servir también la interfaz de usuario en /ui
    [#1752](https://github.com/weaveworks/scope/pull/1752)
*   Nombra nuestras rutas, para que /metrics dé agregaciones más sensatas
    [#1723](https://github.com/weaveworks/scope/pull/1723)
*   Agregar opciones para almacenar informes memcached con diferentes niveles de compresión
    [#1684](https://github.com/weaveworks/scope/pull/1684)

## Versión 0.16.2

Correcciones:

*   Scope no se inicia en nuevas instalaciones de Docker para Mac
    [#1755](https://github.com/weaveworks/scope/issues/1755)

## Versión 0.16.1

Esta es una versión de corrección de errores. Además, la seguridad de la sonda Scope se puede reforzar deshabilitando
controles con el nuevo `--probe.no-controls` , que impide a los usuarios
apertura de terminales, arranque/parada de contenedores, visualización de registros, etc.

Nuevas características y mejoras:

*   Permitir la desactivación de controles en sondeos
    [#1627](https://github.com/weaveworks/scope/pull/1627)
*   Facilitar la desactivación de las integraciones de tejido
    [#1610](https://github.com/weaveworks/scope/pull/1610)
*   Errores de DNS de impresión
    [#1607](https://github.com/weaveworks/scope/pull/1607)
*   Agregue la bandera de ejecución en seco al alcance, para que cuando se lance podamos verificar que los args sean válidos.
    [#1609](https://github.com/weaveworks/scope/pull/1609)

Mejoras de rendimiento:

*   Usar un segmento en lugar de una lista persistente para la acumulación temporal de listas
    [#1660](https://github.com/weaveworks/scope/pull/1660)

Correcciones:

*   Debe comprobar si la sonda ya se está ejecutando cuando se inicia en modo independiente en Docker para Mac
    [#1679](https://github.com/weaveworks/scope/issues/1679)
*   Corrige la posición de las barras de red cuando se selecciona un nodo.
    [#1667](https://github.com/weaveworks/scope/pull/1667)
*   Scope no se inicia en la última versión de Docker para Mac (beta18)
    [#1650](https://github.com/weaveworks/scope/pull/1650)
    [#1669](https://github.com/weaveworks/scope/pull/1669)
*   Corrige el empaquetado de terminales sincronizando los anchos de terminal docker/term.js.
    [#1648](https://github.com/weaveworks/scope/pull/1648)
*   Lado local erróneamente atribuido en las conexiones salientes a Internet
    [#1598](https://github.com/weaveworks/scope/issues/1598)
*   No se puede reenviar el puerto de la aplicación desde kubernetes con el comando en la documentación
    [#1526](https://github.com/weaveworks/scope/issues/1526)
*   Forzar algunos anchos de columna conocidos para evitar el truncamiento de otros
    [#1641](https://github.com/weaveworks/scope/pull/1641)

Documentación:

*   Reemplace wget en las instrucciones con curl, ya que es más ampliamente utilizado. en macs
    [#1670](https://github.com/weaveworks/scope/pull/1670)
*   No antepongas `scope launch` con sudo
    [#1606](https://github.com/weaveworks/scope/pull/1606)
*   Aclarar las instrucciones para usar Scope con Weave Cloud
    [#1611](https://github.com/weaveworks/scope/pull/1611)
*   Página de registro añadida de nuevo
    [#1604](https://github.com/weaveworks/scope/pull/1604)
*   tejer capturas de pantalla en la nube
    [#1603](https://github.com/weaveworks/scope/pull/1603)

Mejoras internas y limpieza:

*   Lint shellscripts de herramientas
    [#1658](https://github.com/weaveworks/scope/pull/1658)
*   Promover fixprobe y eliminar resto de experimentales
    [#1646](https://github.com/weaveworks/scope/pull/1646)
*   refactorizar algunos ayudantes de temporización en un lib común
    [#1642](https://github.com/weaveworks/scope/pull/1642)
*   Ayudante para leer y escribir desde binario
    [#1600](https://github.com/weaveworks/scope/pull/1600)
*   Actualizaciones del documento de proveedores
    [#1595](https://github.com/weaveworks/scope/pull/1595)

Cambios relacionados con Weave Cloud:

*   Almacenar un histograma de los tamaños de los informes
    [#1668](https://github.com/weaveworks/scope/pull/1668)
*   Entrega continua por cable
    [#1654](https://github.com/weaveworks/scope/pull/1654)
*   Cuente las solicitudes de memcache incluso si se agotan el tiempo de espera
    [#1662](https://github.com/weaveworks/scope/pull/1662)
*   Agregar un modo de archivo de informe estático.
    [#1659](https://github.com/weaveworks/scope/pull/1659)
*   Bump memcache expiration
    [#1640](https://github.com/weaveworks/scope/pull/1640)
*   Correcciones para el soporte de memcache
    [#1628](https://github.com/weaveworks/scope/pull/1628)
*   Refactorizar capas de almacenamiento en caché en el colector de dinamo
    [#1616](https://github.com/weaveworks/scope/pull/1616)
*   Reelabore las métricas de alcance de acuerdo con las convenciones de Prometheus.
    [#1615](https://github.com/weaveworks/scope/pull/1615)
*   Corregir el error de puntero nulo cuando memcache no está habilitado
    [#1612](https://github.com/weaveworks/scope/pull/1612)
*   Agregar backoff al cliente cónsul
    [#1608](https://github.com/weaveworks/scope/pull/1608)
*   Query memcached from dynamo db collector
    [#1602](https://github.com/weaveworks/scope/pull/1602)
*   Usar histogramas sobre resúmenes
    [#1665](https://github.com/weaveworks/scope/pull/1665)

## Versión 0.16.0

Resúmenes:

*   Nuevo filtro de red para determinar rápidamente a qué redes pertenecen sus contenedores.

Nuevas características y mejoras:

*   Vista de red
    [#1528](https://github.com/weaveworks/scope/pull/1528)
    [#1593](https://github.com/weaveworks/scope/pull/1593)
*   Nodos de implementación de etiquetas con recuento de réplicas
    [#1530](https://github.com/weaveworks/scope/pull/1530)
*   Agregar indicador para deshabilitar la generación de informes de procesos (y puntos finales supervisados)
    [#1511](https://github.com/weaveworks/scope/pull/1511)
*   Agregar el estado del pod a la tabla de resumen
    [#1523](https://github.com/weaveworks/scope/pull/1523)
*   Agregue filtros para pseudo nodos.
    [#1581](https://github.com/weaveworks/scope/pull/1581)

Mejoras de rendimiento:

*   Inicie rápidamente el ticker de resolución dns para mejorar la latencia del primer informe.
    [#1508](https://github.com/weaveworks/scope/pull/1508)

Correcciones:

*   Arreglar cuadro de búsqueda alto en Firefox
    [#1583](https://github.com/weaveworks/scope/pull/1583)
*   Reportero de sonda atascado
    [#1576](https://github.com/weaveworks/scope/issues/1576)
*   Contenedor en varias redes que no muestra todas las conexiones
    [#1573](https://github.com/weaveworks/scope/issues/1573)
*   La sonda scope se conecta a Localhost & Prod incluso cuando se le dan nombres de host explícitos
    [#1566](https://github.com/weaveworks/scope/issues/1566)
*   Reparar la comprobación de Docker para Mac
    [#1551](https://github.com/weaveworks/scope/pull/1551)
*   Si los objetos k8s solo tienen un contenedor, muestre las métricas de ese contenedor en ellos
    [#1473](https://github.com/weaveworks/scope/pull/1473)
*   Nunca almacenes NUEVOS flujos de conntrack (solo almacenes actualizaciones).
    [#1541](https://github.com/weaveworks/scope/pull/1541)
*   Los pods con > 1 contenedor que realizan conexiones no muestran ninguna conexión
    [#1494](https://github.com/weaveworks/scope/issues/1494)
*   Faltan bordes al usar el controlador IPAM de Docker
    [#1563](https://github.com/weaveworks/scope/issues/1563)
*   Duplicar pila en la vista "por imagen"
    [#1521](https://github.com/weaveworks/scope/issues/1521)

Documentación:

*   Aclarar la coincidencia de versiones de kubectl
    [#1582](https://github.com/weaveworks/scope/pull/1582)
*   Weave Cloud actualizado y configuración aclarada
    [#1586](https://github.com/weaveworks/scope/pull/1586)

Mejoras internas y limpieza:

*   Agregar middleware de identidad
    [#1574](https://github.com/weaveworks/scope/pull/1574)
*   Reescribir net/http. Pedir. {URL. Ruta,RequestURI} de forma coherente
    [#1555](https://github.com/weaveworks/scope/pull/1555)
*   Agregar Marathon JSON para su lanzamiento en el clúster de minimesos
    [#1509](https://github.com/weaveworks/scope/pull/1509)
*   Integración de Circle para la publicación automática de documentos.
    [#1517](https://github.com/weaveworks/scope/pull/1517)
*   Etiquetar imágenes de ámbito en docker hub como lo hacemos en servicio
    [#1572](https://github.com/weaveworks/scope/pull/1572)
*   Ámbito lento: mejorar los mensajes de error para la depuración
    [#1534](https://github.com/weaveworks/scope/pull/1534)
*   circle.yml: implementar ramas que no sean maestras
    [#1535](https://github.com/weaveworks/scope/pull/1535)
*   Agregar insignia de docker hub
    [#1540](https://github.com/weaveworks/scope/pull/1540)
*   Aumentar las réplicas de prueba
    [#1529](https://github.com/weaveworks/scope/pull/1529)
*   Ignorar direcciones IPv6 en Docker Reporter
    [#1552](https://github.com/weaveworks/scope/pull/1552)

Cambios relacionados con Weave Cloud:

*   Agregar encabezado de versión de sondeo a solicitudes de sondeo
    [#1564](https://github.com/weaveworks/scope/pull/1564)
*   Obtener informes no almacenados en caché en paralelo
    [#1554](https://github.com/weaveworks/scope/pull/1554)
*   Varias correcciones para multitenencia
    [#1533](https://github.com/weaveworks/scope/pull/1533)
*   Utilice NATS para los informes de acceso directo en el servicio.
    [#1568](https://github.com/weaveworks/scope/pull/1568)
*   Si no obtenemos un nombre de ruta del enrutador, haga uno desde la url.
    [#1570](https://github.com/weaveworks/scope/pull/1570)
*   Errores de registro en respuesta a solicitudes http.
    [#1569](https://github.com/weaveworks/scope/pull/1569)
*   Poner informes en S3; Agregar almacenamiento en caché de procesos
    [#1545](https://github.com/weaveworks/scope/pull/1545)
*   Utilice la fusión inteligente en el recopilador de DynamoDB.
    [#1543](https://github.com/weaveworks/scope/pull/1543)
*   Permitir que el usuario especifique el nombre de la tabla y el prefijo de la cola.
    [#1538](https://github.com/weaveworks/scope/pull/1538)
*   Obtener el nombre de la ruta antes de la solicitud de munging
    [#1590](https://github.com/weaveworks/scope/pull/1590)

## Versión 0.15.0

Resúmenes:

Esta versión viene con:

*   Búsqueda: nuevo campo de búsqueda inteligente que te permite filtrar lo que puedes ver por
    nombres de contenedores, todo tipo de metadatos, por ejemplo, direcciones IP y métricas
    comparaciones, por ejemplo, CPU > 50%.
*   Visualización mejorada de Kubernetes: filtros de espacio de nombres, ReplicaSet/Deployment
    vistas, metadatos adicionales, mejor navegación, mostrar registros de Pods, eliminar Pods,
    correcciones de errores y más ...
*   Mejoras en el rendimiento de la aplicación Scope: ~ 3 veces más reducción en el consumo de CPU.

Nuevas características y mejoras:

*   Nuevo campo de búsqueda
    [#1429](https://github.com/weaveworks/scope/pull/1429)
    [#1499](https://github.com/weaveworks/scope/pull/1499)
*   Mejoras de Kubernetes:
    *   Vistas Implementación y Conjunto de réplicas
        [#1436](https://github.com/weaveworks/scope/pull/1436)
    *   Agregue controles de escalabilidad vertical/baja en implementaciones, conjuntos de réplicas y controladores de replicación
        [#1451](https://github.com/weaveworks/scope/pull/1451)
    *   Filtrar por espacios de nombres de Kubernetes
        [#1386](https://github.com/weaveworks/scope/pull/1386)
    *   Quitar la restricción de pedidos de implementación de App->Probe
        [#1433](https://github.com/weaveworks/scope/pull/1433)
    *   Mostrar pod IP y # contenedor en la tabla de hijos en el panel de detalles.
        [#1435](https://github.com/weaveworks/scope/pull/1435)
        [#1409](https://github.com/weaveworks/scope/pull/1409)
    *   Agregar controles de eliminación de pod
        [#1368](https://github.com/weaveworks/scope/pull/1368)
    *   Mostrar la IP del equilibrador de carga k8s si está configurada
        [#1378](https://github.com/weaveworks/scope/pull/1378)
    *   Mostrar el número de pods en servicio
        [#1352](https://github.com/weaveworks/scope/pull/1352)
    *   Filtrar contenedores del sistema GKE
        [#1438](https://github.com/weaveworks/scope/pull/1438)
*   Mostrar etiquetas k8s y contenedores env vars en el panel de detalles
    [#1342](https://github.com/weaveworks/scope/pull/1342)
    [#1465](https://github.com/weaveworks/scope/pull/1465)
*   Instrumento `scope help`
    [#1357](https://github.com/weaveworks/scope/pull/1357)
    [#1419](https://github.com/weaveworks/scope/pull/1419)
*   Agregar swarm-agent, swarm-agent master al filtro de contenedor del sistema
    [#1356](https://github.com/weaveworks/scope/pull/1356)
*   Agregue control para eliminar contenedores docker detenidos.
    [#1290](https://github.com/weaveworks/scope/pull/1290)
*   Agregar un botón para descargar el informe como JSON
    [#1365](https://github.com/weaveworks/scope/pull/1365)
*   Utilice la información DNS resuelta inversamente en la tabla de conexiones.
    [#1359](https://github.com/weaveworks/scope/pull/1359)
*   Agregue un nodo 'No administrado' a las vistas k8s que incluían contenedores que no son k8s.
    [#1350](https://github.com/weaveworks/scope/pull/1350)
*   Admite eventos de cambio de nombre de docker
    [#1332](https://github.com/weaveworks/scope/pull/1332)
*   Eliminar la versión de la imagen de los enlaces principales
    [#1348](https://github.com/weaveworks/scope/pull/1348)
*   Agregar compatibilidad con Docker para Mac
    [#1448](https://github.com/weaveworks/scope/pull/1448)

Mejoras de rendimiento:

*   Aplicación Scope:
    *   Una fusión de informe de complejidad logarítmica(n)
        [#1418](https://github.com/weaveworks/scope/pull/1418)
        [#1447](https://github.com/weaveworks/scope/pull/1447)
    *   No combinar nodos en la canalización de representación
        [#1398](https://github.com/weaveworks/scope/pull/1398)
    *   Pase cero para el decorador en la tubería de renderizado cuando sea posible
        [#1397](https://github.com/weaveworks/scope/pull/1397)
*   Sonda de alcance:
    *   Base de precalculación de los nodos contenedores
        [#1456](https://github.com/weaveworks/scope/pull/1456)

Correcciones:

*   Atribuir correctamente las conexiones de corta duración dnaT-ed
    [#1410](https://github.com/weaveworks/scope/pull/1410)
*   No atribuya conexiones conntracked a contenedores de pausa k8s.
    [#1415](https://github.com/weaveworks/scope/pull/1415)
*   No mostrar vistas de kubernetes si no se ejecuta kubernetes
    [#1364](https://github.com/weaveworks/scope/issues/1364)
*   Los nombres de pod que faltan en la vista de pod de kubernetes y los contenedores de pausa no se muestran como hijos de pods
    [#1412](https://github.com/weaveworks/scope/pull/1412)
*   Corregir el recuento de nodos agrupados para los nodos secundarios filtrados
    [#1371](https://github.com/weaveworks/scope/pull/1371)
*   No mostrar etiquetas de contenedores en las imágenes de contenedores
    [#1374](https://github.com/weaveworks/scope/pull/1374)
*   `docker rm -f`Los contenedores ed persisten
    [#1072](https://github.com/weaveworks/scope/issues/1072)
*   De alguna manera, el nodo de Internet se pierde, pero los bordes están ahí
    [#1304](https://github.com/weaveworks/scope/pull/1304)
*   Identificadores de nodo con / conduce a un bucle de redirección cuando el ámbito está montado debajo de una ruta con redirección de barra diagonal
    [#1335](https://github.com/weaveworks/scope/issues/1335)
*   Ignorar las conexiones conntracked en las que nunca vimos una actualización
    [#1466](https://github.com/weaveworks/scope/issues/466)
*   Contenedores atribuidos incorrectamente al host
    [#1472](https://github.com/weaveworks/scope/issues/1472)
*   k8s: Borde inesperado del nodo Internet
    [#1469](https://github.com/weaveworks/scope/issues/1469)
*   Cuando el usuario proporciona IP addr en la línea de comandos, no intentamos conectarnos a localhost
    [#1477](https://github.com/weaveworks/scope/issues/1477)
*   Etiquetas de host incorrectas en los nodos de contenedor
    [#1501](https://github.com/weaveworks/scope/issues/1501)

Documentación:

*   Documentos de ámbito reestructurados
    [#1416](https://github.com/weaveworks/scope/pull/1416)
    [#1479](https://github.com/weaveworks/scope/pull/1479)
*   Agregar instrucciones e insignia de ECS a README
    [#1392](https://github.com/weaveworks/scope/pull/1392)
*   Documentar cómo acceder a la interfaz de usuario de ámbito en k8s
    [#1426](https://github.com/weaveworks/scope/pull/1426)
*   Actualizar el archivo Léame para expresar que los conjuntos de demonios no se programarán en nodos no escadulables antes de kubernetes 1.2
    [#1434](https://github.com/weaveworks/scope/pull/1434)

Mejoras internas y limpieza:

*   Migrar de Flux a Redux
    [#1388](https://github.com/weaveworks/scope/pull/1388)
*   Agregar indicador de punto de control de kubernetes
    [#1391](https://github.com/weaveworks/scope/pull/1391)
*   Agregar middleware de reescritura de ruta genérica
    [#1381](https://github.com/weaveworks/scope/pull/1381)
*   Informe del nombre de host y la versión en la estructura de la sonda y la versión en el nodo host.
    [#1377](https://github.com/weaveworks/scope/pull/1377)
*   Reorganizar el render/paquete
    [#1360](https://github.com/weaveworks/scope/pull/1360)
*   Huellas dactilares de activos
    [#1354](https://github.com/weaveworks/scope/pull/1354)
*   Actualizar a go1.6.2
    [#1362](https://github.com/weaveworks/scope/pull/1362)
*   Agregue búfer al canal mockPublisher para evitar el interbloqueo entre Publish() y Stop()
    [#1358](https://github.com/weaveworks/scope/pull/1358)
*   Agregar un resumen explícito del nodo de grupo en lugar de hacerlo en los otros resúmenes
    [#1327](https://github.com/weaveworks/scope/pull/1327)
*   Ya no cree códecs para render / paquete.
    [#1345](https://github.com/weaveworks/scope/pull/1345)
*   Medir el tamaño de los informes
    [#1458](https://github.com/weaveworks/scope/pull/1458)

## Versión 0.14.0

Resúmenes:

Esta versión viene con dos nuevas características principales.

*   Complementos de sonda: ahora puede crear su complemento basado en HTTP para proporcionar nuevas métricas
    y mostrarlos en Scope. Puedes leer más al respecto y ver algunos ejemplos
    [aquí](https://github.com/weaveworks/scope/tree/master/examples/plugins).
*   Métricas en el lienzo: las métricas ahora se muestran en los nodos y no solo en
    el panel de detalles, comenzando con el consumo de CPU y memoria.

Además, el rendimiento de la interfaz de usuario se ha mejorado considerablemente y los 100 nodos
se ha levantado el límite de renderizado.

Nuevas características y mejoras:

*   Complementos de sonda
    [#1126](https://github.com/weaveworks/scope/pull/1126)
    [#1277](https://github.com/weaveworks/scope/pull/1277)
    [#1280](https://github.com/weaveworks/scope/pull/1280)
    [#1283](https://github.com/weaveworks/scope/pull/1283)
*   Métricas en lienzo
    [#1105](https://github.com/weaveworks/scope/pull/1105)
    [#1204](https://github.com/weaveworks/scope/pull/1204)
    [#1225](https://github.com/weaveworks/scope/pull/1225)
    [#1243](https://github.com/weaveworks/scope/issues/1243)
*   Mejoras en el panel detalles del nodo
    *   Agregar tablas de conexión
        [#1017](https://github.com/weaveworks/scope/pull/1017)
        [#1248](https://github.com/weaveworks/scope/pull/1248)
    *   Diseño: hacer un mejor uso del espacio de la columna
        [#1272](https://github.com/weaveworks/scope/pull/1272)
    *   Minigráficos
        *   Actualiza cada segundo y muestra el historial de 60 segundos
            [#795](https://github.com/weaveworks/scope/pull/795)
        *   Aplicar formato a la información sobre herramientas en los desplazamientos
            [#1230](https://github.com/weaveworks/scope/pull/1230)
    *   Ordenar las entradas numéricas (por ejemplo, recuentos de imágenes, id. de proceso) como se esperaba
        [#1125](https://github.com/weaveworks/scope/pull/1125)
    *   Eliminar métricas load5 y load15
        [#1274](https://github.com/weaveworks/scope/pull/1274)
*   Mejoras en la vista de gráficos
    *   Mejoras en el filtrado de nodos
        *   Introducir selectores de filtrado de tres vías (por ejemplo, elegir entre *Contenedores del sistema*, *Contenedores de aplicaciones* o *Ambos*)
            [#1159](https://github.com/weaveworks/scope/pull/1159)
        *   Mantener la selección de filtrado de nodos en todas las subvistas (por ejemplo, *Contenedores por ID* y *Contenedores por imagen*)
            [#1237](https://github.com/weaveworks/scope/pull/1237)
    *   Refinar la longitud máxima de los nombres de nodo
        [#1263](https://github.com/weaveworks/scope/issues/1263)
        [#1255](https://github.com/weaveworks/scope/pull/1255)
    *   Refinar el ancho de borde de los nodos
        [#1138](https://github.com/weaveworks/scope/pull/1138)
        [#1120](https://github.com/weaveworks/scope/pull/1120)
    *   Panorámica/zoom de caché por topología
        [#1261](https://github.com/weaveworks/scope/pull/1261)
*   Habilitar el inicio de terminales en hosts
    [#1208](https://github.com/weaveworks/scope/pull/1208)
*   Permitir pausar la interfaz de usuario a través de un botón
    [#1106](https://github.com/weaveworks/scope/pull/1106)
*   Divida el nodo de Internet para las conexiones entrantes y salientes.
    [#566](https://github.com/weaveworks/scope/pull/566)
*   Mostrar el estado del pod k8s
    [#1289](https://github.com/weaveworks/scope/pull/1289)
*   Permitir la personalización del nombre de host de Scope en Weave Net con `scope launch --weave.hostname`
    [#1041](https://github.com/weaveworks/scope/pull/1041)
*   Rebautizar `--weave.router.addr` Para `--weave.addr` en la sonda para comprobar la coherencia con la aplicación
    [#1060](https://github.com/weaveworks/scope/issues/1060)
*   Soporte nuevo `sha256:` Identificadores de imagen de Docker
    [#1161](https://github.com/weaveworks/scope/pull/1161)
    [#1184](https://github.com/weaveworks/scope/pull/1184)
*   Controlar las desconexiones del servidor correctamente en la interfaz de usuario
    [#1140](https://github.com/weaveworks/scope/pull/1140)

Mejoras de rendimiento:

*   Mejoras de rendimiento para el lienzo de la interfaz de usuario
    [#1186](https://github.com/weaveworks/scope/pull/1186)
    [#1236](https://github.com/weaveworks/scope/pull/1236)
    [#1239](https://github.com/weaveworks/scope/pull/1239)
    [#1262](https://github.com/weaveworks/scope/pull/1262)
    [#1259](https://github.com/weaveworks/scope/pull/1259)
*   Reducir el consumo de CPU si la interfaz de usuario no puede conectarse al back-end
    [#1229](https://github.com/weaveworks/scope/pull/1229)

Correcciones:

*   La aplicación Scope no caduca correctamente los informes antiguos
    [#1286](https://github.com/weaveworks/scope/issues/1286)
*   Los nodos de contenedor aparecen sin una etiqueta de host
    [#1065](https://github.com/weaveworks/scope/issues/1065)
*   Cambiar el tamaño de la ventana y acercar/alejar puede confundir el tamaño de la ventana
    [#1180](https://github.com/weaveworks/scope/issues/1096)
*   Enlace desde contenedor -> Pod no funciona
    [#1180](https://github.com/weaveworks/scope/issues/1293)
*   Varias correcciones de calcetines y tuberías.
    [#1172](https://github.com/weaveworks/scope/pull/1172)
    [#1175](https://github.com/weaveworks/scope/pull/1175)
*   Hacer `--app-only` Solo ejecute la aplicación y no la sonde
    [#1067](https://github.com/weaveworks/scope/pull/1067)
*   El archivo SVG exportado arroja el error "CANT" en Adobe Illustrator
    [#1144](https://github.com/weaveworks/scope/issues/1144)
*   Las etiquetas de Docker no se representan correctamente
    [#1284](https://github.com/weaveworks/scope/issues/1284)
*   Error al analizar la versión del kernel en `/proc` lector de fondo
    [#1136](https://github.com/weaveworks/scope/issues/1136)
*   Abrir la terminal no abre el trabajo para algunos contenedores
    [#1195](https://github.com/weaveworks/scope/issues/1195)
*   Terminales: Intente averiguar qué shell usar en lugar de simplemente ejecutar `/bin/sh`
    [#1069](https://github.com/weaveworks/scope/pull/1069)
*   Corregir el tamaño del logotipo incrustado para Safari
    [#1084](https://github.com/weaveworks/scope/pull/1084)
*   No leas desde la aplicación. Versión antes de inicializarla
    [#1163](https://github.com/weaveworks/scope/pull/1163)
*   No mostrar varios pseudonodos en la vista de host para la misma IP
    [#1155](https://github.com/weaveworks/scope/issues/1155)
*   Arreglar las condiciones de carrera detectadas por el detector de carrera de Go 1.6
    [#1192](https://github.com/weaveworks/scope/issues/1192)
    [#1087](https://github.com/weaveworks/scope/issues/1087)

Documentación:

*   Proporcionar ejemplos de Docker Compose para iniciar la sonda Scope con Scope Cloud Service
    [#1146](https://github.com/weaveworks/scope/pull/1146)

Características experimentales:

*   Demostración de actualización para el trazador
    [#1157](https://github.com/weaveworks/scope/pull/1157)

Cambios relacionados con el modo de servicio:

*   Agregar `/api/probes` Extremo
    [#1265](https://github.com/weaveworks/scope/pull/1265)
*   Mejoras en el soporte multitenencia
    [#996](https://github.com/weaveworks/scope/pull/996)
    [#1150](https://github.com/weaveworks/scope/pull/1150)
    [#1200](https://github.com/weaveworks/scope/pull/1200)
    [#1241](https://github.com/weaveworks/scope/pull/1241)
    [#1209](https://github.com/weaveworks/scope/pull/1209)
    [#1232](https://github.com/weaveworks/scope/pull/1232)

Mejoras internas y limpieza:

*   Hacer que los objetos resaltadores de nodo/borde sean inmutables en la tienda de aplicaciones
    [#1173](https://github.com/weaveworks/scope/pull/1173)
*   Haga que el procesamiento perimetral en caché sea más robusto
    [#1254](https://github.com/weaveworks/scope/pull/1254)
*   Hacer que el objeto de topologías de la tienda de aplicaciones sea inmutable
    [#1167](https://github.com/weaveworks/scope/pull/1167)
*   Prueba Fix TestCollector
    [#1070](https://github.com/weaveworks/scope/pull/1070)
*   Actualizar el cliente de Docker para obtener mejores cadenas de estado en la interfaz de usuario
    [#1235](https://github.com/weaveworks/scope/pull/1235)
*   Actualizar a go1.6
    [#1077](https://github.com/weaveworks/scope/pull/1077)
*   Actualizaciones de React/lodash/babel + linting actualizado (linted)
    [#1171](https://github.com/weaveworks/scope/pull/1171)
*   Quitar topología de direcciones
    [#1127](https://github.com/weaveworks/scope/pull/1127)
*   Agregar documentos de proveedores
    [#1180](https://github.com/weaveworks/scope/pull/1180)
*   Arreglar hacer que el cliente-inicio
    [#1210](https://github.com/weaveworks/scope/pull/1210)
*   Degradar react-motion
    [#1183](https://github.com/weaveworks/scope/pull/1183)
*   Haz que bin/release funcione en un mac.
    [#887](https://github.com/weaveworks/scope/pull/887)
*   Agregue varios middleware a la aplicación.
    [#1234](https://github.com/weaveworks/scope/pull/1234)
*   Hacer que la compilación no conteinerizada funcione en OSX
    [#1028](https://github.com/weaveworks/scope/pull/1028)
*   Eliminar el archivo generado por codecgen antes de crear el paquete
    [#1135](https://github.com/weaveworks/scope/pull/1135)
*   Compilar/instalar paquetes antes de invocar codecgen
    [#1042](https://github.com/weaveworks/scope/pull/1042)
*   circle.yml: agregar variable $DOCKER_ORGANIZATION
    [#1083](https://github.com/weaveworks/scope/pull/1083)
*   circle.yml: implementar en una cuenta de concentrador personal
    [#1055](https://github.com/weaveworks/scope/pull/1055)
*   circle.yml: deshabilite las compilaciones GCE cuando falten credenciales
    [#1054](https://github.com/weaveworks/scope/pull/1054)
*   Limpie todo el JS en el directorio de compilación del cliente.
    [#1205](https://github.com/weaveworks/scope/pull/1205)
*   Elimine los archivos temporales en el contenedor de compilación para reducirlo en ~ 100 MB
    [#1206](https://github.com/weaveworks/scope/pull/1206)
*   Actualizar herramientas y crear contenedor para comprobar si hay errores ortográficos
    [#1199](https://github.com/weaveworks/scope/pull/1199)
*   Solucione un par de problemas menores para goreportcard y agregue una insignia para ello.
    [#1203](https://github.com/weaveworks/scope/pull/1203)

## Versión 0.13.1

Correcciones:

*   Hacer que las tuberías funcionen con scope.weave.works
    [#1099](https://github.com/weaveworks/scope/pull/1099)
    [#1085](https://github.com/weaveworks/scope/pull/1085)
    [#994](https://github.com/weaveworks/scope/pull/994)
*   No entre en pánico al verificar que la versión falla
    [#1117](https://github.com/weaveworks/scope/pull/1117)

## Versión 0.13.0

Nota: Esta versión viene con grandes mejoras de rendimiento, reduciendo el uso de la CPU de la sonda en un 70% y el uso de la CPU de la aplicación hasta en un 85%. Vea los cambios detallados relacionados con la mejora del rendimiento a continuación:

Mejoras de rendimiento:

*   Mejore el rendimiento del códec
    [#916](https://github.com/weaveworks/scope/pull/916)
    [#1002](https://github.com/weaveworks/scope/pull/1002)
    [#1005](https://github.com/weaveworks/scope/pull/1005)
    [#980](https://github.com/weaveworks/scope/pull/980)
*   Reducir la cantidad de objetos asignados por el códec
    [#1000](https://github.com/weaveworks/scope/pull/1000)
*   Aplicación Refactor para multitenencia
    [#997](https://github.com/weaveworks/scope/pull/997)
*   Mejore el rendimiento de la obtención de estadísticas de Docker
    [#989](https://github.com/weaveworks/scope/pull/989)
*   Límite de velocidad de lectura de archivos proc
    [#912](https://github.com/weaveworks/scope/pull/912)
    [#905](https://github.com/weaveworks/scope/pull/905)
*   Compilar selectores k8s una vez (no para cada pod)
    [#918](https://github.com/weaveworks/scope/pull/918)
*   Corregir la lectura de inodos de espacio de nombres de red
    [#898](https://github.com/weaveworks/scope/pull/898)

Nuevas características y mejoras:

*   Formas de nodo para diferentes topologías, por ejemplo, heptágonos para pods de Kubernetes
    [#884](https://github.com/weaveworks/scope/pull/884)
    [#1006](https://github.com/weaveworks/scope/pull/1006)
    [#1037](https://github.com/weaveworks/scope/pull/1037)
*   Botón de retransmisión forzada que puede ayudar con los diseños de topología que tienen muchos cruces de bordes
    [#981](https://github.com/weaveworks/scope/pull/981)
*   Botón Descargar para guardar el gráfico de nodos actual como archivo SVG
    [#1027](https://github.com/weaveworks/scope/pull/1027)
*   Reemplace los botones Mostrar más con intercalaciones con recuentos
    [#1012](https://github.com/weaveworks/scope/pull/1012)
    [#1029](https://github.com/weaveworks/scope/pull/1029)
*   Mejorar el contraste de la vista predeterminada
    [#979](https://github.com/weaveworks/scope/pull/979)
*   Botón de modo de alto contraste para ver el alcance en proyectores
    [#954](https://github.com/weaveworks/scope/pull/954)
    [#984](https://github.com/weaveworks/scope/pull/984)
*   Recopilar descriptores de archivos como métrica de proceso
    [#961](https://github.com/weaveworks/scope/pull/961)
*   Mostrar etiquetas de Docker en su propia tabla en el panel de detalles
    [#904](https://github.com/weaveworks/scope/pull/904)
    [#965](https://github.com/weaveworks/scope/pull/965)
*   Mejorar el resaltado de la topología seleccionada
    [#936](https://github.com/weaveworks/scope/pull/936)
    [#964](https://github.com/weaveworks/scope/pull/964)
*   Detalles: solo muestra metadatos importantes de forma predeterminada, expande el resto
    [#946](https://github.com/weaveworks/scope/pull/946)
*   Reordenar las tablas secundarias en el panel de detalles por importancia
    [#941](https://github.com/weaveworks/scope/pull/941)
*   Acorte los identificadores de contenedor e imagen de Docker en el panel de detalles.
    [#930](https://github.com/weaveworks/scope/pull/930)
*   Acortar algunas etiquetas de panel de detalles que se truncaron
    [#940](https://github.com/weaveworks/scope/pull/940)
*   Agregar columna Recuento de contenedores a la tabla de imágenes de contenedor
    [#919](https://github.com/weaveworks/scope/pull/919)
*   Compruebe periódicamente si hay versiones más recientes del ámbito.
    [#907](https://github.com/weaveworks/scope/pull/907)
*   Cambie el nombre de Aplicaciones -> Proceso, ordene las topologías por rango.
    [#866](https://github.com/weaveworks/scope/pull/866)
*   Cambie el nombre 'por nombre de host' a 'por nombre dns'
    [#856](https://github.com/weaveworks/scope/pull/856)
*   Agregue el tiempo de actividad del contenedor y reinicie el recuento en el panel de detalles.
    [#853](https://github.com/weaveworks/scope/pull/853)
*   Utilice las instrucciones de conexión de conntrack para mejorar el flujo de diseño
    [#967](https://github.com/weaveworks/scope/pull/967)
*   Compatibilidad con controles de contenedores en Kubernetes
    [#1043](https://github.com/weaveworks/scope/pull/1043)
*   Agregar registro de depuración
    [#935](https://github.com/weaveworks/scope/pull/935)

Correcciones:

*   Usar TCP para tejer dns para corregir el autoclustering
    [#1038](https://github.com/weaveworks/scope/pull/1038)
*   Agregue ping/pong al protocolo websocket para evitar que las conexiones websocket se caigan al atravesar los balancers de carga
    [#995](https://github.com/weaveworks/scope/pull/995)
*   Controlar el cierre del canal de eventos de Docker con gracia
    [#1014](https://github.com/weaveworks/scope/pull/1014)
*   No mostrar la fila de metadatos de DIRECCIONES IP en blanco para contenedores sin IP
    [#960](https://github.com/weaveworks/scope/pull/960)
*   Eliminar las matemáticas del puntero (comparación) del almacenamiento en caché de renderizado, ya que no es confiable
    [#962](https://github.com/weaveworks/scope/pull/962)
*   Establezca TERM=XTerm en execs para evitar el problema 9299 de Docker
    [#969](https://github.com/weaveworks/scope/pull/969)
*   Corregir el bloqueo del etiquetador de tejido
    [#976](https://github.com/weaveworks/scope/pull/976)
*   Utilice el registrador Sirupsen/logrus en el etiquetador Weave
    [#974](https://github.com/weaveworks/scope/pull/974)
*   Corregir la codificación JSON para fixedprobe
    [#975](https://github.com/weaveworks/scope/pull/975)
*   No representar ninguna métrica/metadato para el nodo no contenido
    [#956](https://github.com/weaveworks/scope/pull/956)
*   Actualice go-dockerclient para corregir errores con docker 1.10
    [#952](https://github.com/weaveworks/scope/pull/952)
*   Mostrar buenas etiquetas de columna cuando ningún niño tiene métricas
    [#950](https://github.com/weaveworks/scope/pull/950)
*   Corrige el diseño proceso por nombre con los nodos ./foo y /foo
    [#948](https://github.com/weaveworks/scope/pull/948)
*   Lidiar con el tejido de inicio / detención mientras el alcance se está ejecutando
    [#867](https://github.com/weaveworks/scope/pull/867)
*   Eliminar enlaces de host que se vinculan a sí mismos en el panel de detalles
    [#917](https://github.com/weaveworks/scope/pull/917)
*   Simplemente muestre la etiqueta no autenticada en la información sobre herramientas en niños
    [#911](https://github.com/weaveworks/scope/pull/911)
*   Tomar un bloqueo de lectura dos veces solo funciona la mayor parte del tiempo.
    [#889](https://github.com/weaveworks/scope/pull/889)
*   El encabezado de la tabla del panel Detalles busca la etiqueta en todas las filas
    [#895](https://github.com/weaveworks/scope/pull/895)
*   Corrige algunos campos que se desbordan mal en el panel de detalles de Chrome 48
    [#892](https://github.com/weaveworks/scope/pull/892)
*   Evite que aparezcan tarjetas de detalles sobre el terminal.
    [#882](https://github.com/weaveworks/scope/pull/882)
*   Corrige la falta de coincidencia de color del nodo host/panel de detalles
    [#880](https://github.com/weaveworks/scope/pull/880)
*   No registre los errores esperados de websocket
    [#1024](https://github.com/weaveworks/scope/pull/1024)
*   Sobrescribir /etc/weave/apps, porque es posible que ya exista
    [#959](https://github.com/weaveworks/scope/pull/959)
*   Registrar una advertencia cuando los reporteros o etiquetadores tardan demasiado en generarse
    [#944](https://github.com/weaveworks/scope/pull/944)
*   Refactorización menor de metadatos backend y representación de métricas
    [#920](https://github.com/weaveworks/scope/pull/920)
*   Agregue algunas pruebas y un valor cero para el informe. Establece
    [#903](https://github.com/weaveworks/scope/pull/903)

Cree mejoras y limpieza:

*   Deshabilite el punto de control en las pruebas.
    [#1031](https://github.com/weaveworks/scope/pull/1031)
*   Desactive GC para compilaciones.
    [#1023](https://github.com/weaveworks/scope/pull/1023)
*   Nombre de la plantilla Bump para obtener la última versión de docker.
    [#998](https://github.com/weaveworks/scope/pull/998)
*   Corrige el ámbito de construcción fuera de un contenedor.
    [#901](https://github.com/weaveworks/scope/pull/901)
*   No necesita sudo cuando DOCKER_HOST es tcp.
    [#888](https://github.com/weaveworks/scope/pull/888)
*   Deshabilitar el progreso de npm para acelerar la compilación
    [#894](https://github.com/weaveworks/scope/pull/894)
*   Refactorización de deepequal para satisfacer linter
    [#890](https://github.com/weaveworks/scope/pull/890)

Documentación:

*   Documentar cómo obtener perfiles sin `go tool pprof`
    [#993](https://github.com/weaveworks/scope/pull/993)
*   Usar URL corta para la descarga del ámbito
    [#1018](https://github.com/weaveworks/scope/pull/1018)
*   Se ha agregado una nota sobre la dependencia de docker y go al archivo Léame
    [#966](https://github.com/weaveworks/scope/pull/966)
*   Actualizar léame e imágenes.
    [#885](https://github.com/weaveworks/scope/pull/885)
*   Enfoque de actualización para activar volcados de señal
    [#883](https://github.com/weaveworks/scope/pull/883)

## Versión 0.12.0

Nuevas características y mejoras:

*   Nuevo panel interactivo de detalles contextuales
    [#752](https://github.com/weaveworks/scope/pull/752)
*   Recopilar métricas de CPU y memoria por proceso
    [#767](https://github.com/weaveworks/scope/pull/767)
*   k8s: Usar el token de cuenta de servicio de forma predeterminada y mejorar el registro de errores
    [#808](https://github.com/weaveworks/scope/pull/808)
*   k8s: Filtrar la pausa como un contenedor del sistema para ordenar la vista
    [#823](https://github.com/weaveworks/scope/pull/823)
*   k8s: Representar nombres de contenedores a partir de la etiqueta "io.kubernetes.container.name"
    [#810](https://github.com/weaveworks/scope/pull/810)
*   Los sondeos ahora usan TLS en scope.weave.works de forma predeterminada
    [#785](https://github.com/weaveworks/scope/pull/785)
*   Permitir el descarte de un terminal desconectado con \<esc>
    [#819](https://github.com/weaveworks/scope/pull/819)

Correcciones:

*   Correcciones generales de k8s
    [#834](https://github.com/weaveworks/scope/pull/834)
*   Utilice argv\[0] para el nombre del proceso, la aplicación de ámbito diferenciado y la sonda.
    [#796](https://github.com/weaveworks/scope/pull/796)
*   No entre en pánico si no entiende el mensaje en el control WS.
    [#793](https://github.com/weaveworks/scope/pull/793)
*   Resalte un solo nodo no conectado al desplazarse.
    [#790](https://github.com/weaveworks/scope/pull/790)
*   Correcciones al cambio de tamaño de la terminal y al soporte clave
    [#766](https://github.com/weaveworks/scope/pull/766)
    [#780](https://github.com/weaveworks/scope/pull/780)
    [#817](https://github.com/weaveworks/scope/pull/817)
*   Contraer correctamente los nodos en la vista Imágenes de contenedor cuando utilizan un puerto no estándar.
    [#824](https://github.com/weaveworks/scope/pull/824)
*   Dejar de bloquear el alcance del cromo cuando obtenemos bordes "largos".
    [#837](https://github.com/weaveworks/scope/pull/837)
*   Corregir los controles de nodo para que se comporten de forma independiente entre los nodos
    [#797](https://github.com/weaveworks/scope/pull/797)

Cree mejoras y limpieza:

*   Actualice a la última versión de tools.git
    [#816](https://github.com/weaveworks/scope/pull/816)
*   Actualización a la última versión de go-dockerclient
    [#788](https://github.com/weaveworks/scope/pull/788)
*   Acelerar las acumulaciones
    [#775](https://github.com/weaveworks/scope/pull/775)
    [#789](https://github.com/weaveworks/scope/pull/789)
*   Acelerar las pruebas
    [#799](https://github.com/weaveworks/scope/pull/799)
    [#807](https://github.com/weaveworks/scope/pull/807)
*   Divida y mueva el paquete xfer.
    [#794](https://github.com/weaveworks/scope/pull/794)
*   Agregar más pruebas a procspy
    [#751](https://github.com/weaveworks/scope/pull/751)
    [#781](https://github.com/weaveworks/scope/pull/781)
*   Cree una aplicación de ejemplo en el contenedor.
    [#831](https://github.com/weaveworks/scope/pull/831)
*   Varias mejoras para construir y probar
    [#829](https://github.com/weaveworks/scope/pull/829)

## Versión 0.11.1

Corrección de errores:

*   Raspar /proc/PID/net/tcp6 de tal manera que veamos ambos extremos de las conexiones locales
    [cambio](https://github.com/weaveworks/scope/commit/550f21511a2da20717c6de6172b5bf2e9841d905)

## Versión 0.11.0

Nuevas características:

*   Agregar un terminal a la interfaz de usuario con la capacidad de `attach` a, o `exec` un shell in, un contenedor Docker.
    [#650](https://github.com/weaveworks/scope/pull/650)
    [#735](https://github.com/weaveworks/scope/pull/735)
    [#726](https://github.com/weaveworks/scope/pull/726)
*   Añadido `scope version` mandar
    [#750](https://github.com/weaveworks/scope/pull/750)
*   Varias reducciones de uso de CPU para sonda
    [#742](https://github.com/weaveworks/scope/pull/742)
    [#741](https://github.com/weaveworks/scope/pull/741)
    [#737](https://github.com/weaveworks/scope/pull/737)
*   Mostrar el nombre de host de la aplicación a la que está conectado en la parte inferior derecha de la interfaz de usuario
    [#707](https://github.com/weaveworks/scope/pull/707)
*   Agregar métricas de uso de CPU y memoria de host al panel de detalles
    [#711](https://github.com/weaveworks/scope/pull/711)
*   Agregar compatibilidad con json a la aplicación POST /api/report
    [#722](https://github.com/weaveworks/scope/pull/722)
*   Actualice la versión de Docker que incrustamos en la imagen de ámbito a 1.6.2 en sincronía con los cambios de tejido 1.3.
    [#702](https://github.com/weaveworks/scope/pull/702)
*   Mostrar un spinner mientras se cargan los detalles del nodo
    [#691](https://github.com/weaveworks/scope/pull/691)
*   Coloración determinista de nodos basada en rango y etiqueta
    [#694](https://github.com/weaveworks/scope/pull/694)

Correcciones:

*   Mitigar los diseños de una línea de nodos (cuando el gráfico tiene pocas conexiones), el diseño en rectángulo en su lugar
    [#679](https://github.com/weaveworks/scope/pull/679)
*   Al filtrar nodos no conectados en la vista de procesos, filtre también los nodos que solo están conectados a sí mismos.
    [#706](https://github.com/weaveworks/scope/pull/706)
*   Ocultar correctamente el contenedor en función de las etiquetas docker en la imagen del contenedor.
    [#705](https://github.com/weaveworks/scope/pull/705)
*   Muestre los detalles del contenedor detenido en la vista predeterminada, al no aplicar filtros a los extremos de detalles del nodo.
    [#704](https://github.com/weaveworks/scope/pull/704)
    [#701](https://github.com/weaveworks/scope/pull/701)
*   Solucionar problemas de renderizado en Safari
    [#686](https://github.com/weaveworks/scope/pull/686)
*   Tomar la opción de topología predeterminada si falta en la URL
    [#678](https://github.com/weaveworks/scope/pull/678)
*   No trate el nodo faltante como un error de interfaz de usuario
    [#677](https://github.com/weaveworks/scope/pull/677)
*   Anular el establecimiento de detalles anteriores al anular la selección de un nodo
    [#675](https://github.com/weaveworks/scope/pull/675)
*   Agregar x para cerrar el panel de detalles de nuevo
    [#673](https://github.com/weaveworks/scope/pull/673)

Documentación:

*   Agregue una advertencia de seguridad básica.
    [#703](https://github.com/weaveworks/scope/pull/703)
*   Agregar procedimientos básicos de kubernetes al archivo Léame
    [#669](https://github.com/weaveworks/scope/pull/669)
*   Opciones de depuración de documentos para desarrolladores
    [#723](https://github.com/weaveworks/scope/pull/723)
*   Agregue la sección 'obtener ayuda' y actualice la captura de pantalla
    [#709](https://github.com/weaveworks/scope/pull/709)

Cree mejoras y limpieza:

*   No vayas a buscar tejido, git clónalo para que los errores de compilación del tejido no afecten al alcance.
    [#743](https://github.com/weaveworks/scope/pull/743)
*   Reduzca el tamaño de la imagen y el tiempo de compilación mediante la combinación de la sonda de ámbito y los archivos binarios de la aplicación.
    [#732](https://github.com/weaveworks/scope/pull/732)
*   Limpieza de código muerto alrededor de los bordes y edgemetadata
    [#730](https://github.com/weaveworks/scope/pull/730)
*   Hacer `make` Compilar la interfaz de usuario
    [#728](https://github.com/weaveworks/scope/pull/728)
*   Omita el campo de controles de json si está vacío.
    [#725](https://github.com/weaveworks/scope/pull/725)
*   JS a ES2015
    [#712](https://github.com/weaveworks/scope/pull/712)
*   Reacción actualizada a 0.14.3
    [#687](https://github.com/weaveworks/scope/pull/687)
*   Tabla de detalles de nodo limpiada
    [#676](https://github.com/weaveworks/scope/pull/676)
*   Corregir la advertencia de clave de reacción
    [#672](https://github.com/weaveworks/scope/pull/672)

## Versión 0.10.0

Notas:

*   Debido a que la interfaz de usuario de scope ahora puede iniciar/detener/reiniciar Docker
    contenedores, no es prudente tenerlo accesible para personas que no son de confianza
    Partes.

Nuevas características:

*   Agregar controles de ciclo de vida (inicio/parada/reinicio) para contenedores docker
    [#598](https://github.com/weaveworks/scope/pull/598)
    [#642](https://github.com/weaveworks/scope/pull/642)
*   Agregar minigráficos a la interfaz de usuario para algunas métricas
    [#622](https://github.com/weaveworks/scope/pull/622)
*   Mostrar un mensaje cuando la topología seleccionada está vacía
    [#505](https://github.com/weaveworks/scope/pull/505)

Correcciones:

*   Cambiar el diseño de los nodos de forma incremental para reducir los rediseños
    [#593](https://github.com/weaveworks/scope/pull/593)
*   Mejorar la capacidad de respuesta de las actualizaciones de la interfaz de usuario a los cambios de estado del contenedor
    [#628](https://github.com/weaveworks/scope/pull/628)
    [#640](https://github.com/weaveworks/scope/pull/640)
*   Controlar la resolución DNS en un conjunto de nombres
    [#639](https://github.com/weaveworks/scope/pull/639)
*   Mostrar correctamente los recuentos de nodos para subtopologías
    [#621](https://github.com/weaveworks/scope/issues/621)
*   Permitir que el ámbito se inicie después de actualizarse
    [#617](https://github.com/weaveworks/scope/pull/617)
*   Evitar que aparezcan pseudonodos varados en la vista de contenedor
    [#627](https://github.com/weaveworks/scope/pull/627)
    [#674](https://github.com/weaveworks/scope/pull/674)
*   Paralelizar y mejorar la infraestructura de pruebas
    [#614](https://github.com/weaveworks/scope/pull/614)
    [#618](https://github.com/weaveworks/scope/pull/618)
    [#644](https://github.com/weaveworks/scope/pull/644)

## Versión 0.9.0

Nuevas características:

*   Agregar vistas básicas de Kubernetes para pods y servicios
    [#441](https://github.com/weaveworks/scope/pull/441)
*   Soporte para Weave 1.2
    [#574](https://github.com/weaveworks/scope/pull/574)
*   Vista Agregar contenedores por nombre de host
    [#545](https://github.com/weaveworks/scope/pull/545)
*   Cree con Go 1.5, con dependencias de proveedores
    [#584](https://github.com/weaveworks/scope/pull/584)
*   Hacer `scope launch` trabajar desde hosts remotos, con un DOCKER_HOST definido adecuadamente
    [#524](https://github.com/weaveworks/scope/pull/524)
*   Aumente la frecuencia de sondeo de DNS de modo que los clústeres de ámbito se acumulen más rápidamente
    [#524](https://github.com/weaveworks/scope/pull/524)
*   Agregar `scope command` Para imprimir los comandos de Docker utilizados para ejecutar Scope
    [#553](https://github.com/weaveworks/scope/pull/553)
*   Incluir documentación básica sobre cómo ejecutar Scope
    [#572](https://github.com/weaveworks/scope/pull/572)
*   Advertir si el usuario intenta ejecutar Scope en las versiones de Docker <1.5.0
    [#557](https://github.com/weaveworks/scope/pull/557)
*   Agregar compatibilidad para cargar la interfaz de usuario de ámbito desde puntos de conexión https
    [#572](https://github.com/weaveworks/scope/pull/572)
*   Agregar compatibilidad con el envío de informes de sondeos a puntos de conexión https
    [#575](https://github.com/weaveworks/scope/pull/575)

Correcciones:

*   Rastree correctamente las conexiones de corta duración desde Internet
    [#493](https://github.com/weaveworks/scope/pull/493)
*   Corregir un caso de esquina en el que las conexiones de corta duración entre contenedores se atribuyen incorrectamente
    [#577](https://github.com/weaveworks/scope/pull/577)
*   Asegúrese de que se envían las credenciales de servicio al realizar el protocolo de enlace inicial de sondeo<->aplicación
    [#564](https://github.com/weaveworks/scope/pull/564)
*   Ordenar nombres resueltos por DNS inverso para mitigar algunos aleteos de la interfaz de usuario
    [#562](https://github.com/weaveworks/scope/pull/562)
*   No filtre gortulinas en la sonda
    [#531](https://github.com/weaveworks/scope/issues/531)
*   Vuelva a ejecutar los procesos de seguimiento en segundo plano si fallan
    [#581](https://github.com/weaveworks/scope/issues/581)
*   Cree y pruebe con Go 1.5 y proporcione todas las dependencias
    [#584](https://github.com/weaveworks/scope/pull/584)
*   Corregir el error "cerrar en canal nulo" al apagar
    [#599](https://github.com/weaveworks/scope/issues/599)

## Versión 0.8.0

Nuevas características:

*   Mostrar mensaje en la interfaz de usuario cuando las topologías superan los límites de tamaño.
    [#474](https://github.com/weaveworks/scope/issues/474)
*   Proporcione información de imagen de contenedor en el panel de detalles para contenedores.
    [#398](https://github.com/weaveworks/scope/issues/398)
*   Al filtrar los contenedores del sistema, filtre también los pseudonodos, si solo estaban conectados a los contenedores del sistema.
    [#483](https://github.com/weaveworks/scope/issues/483)
*   Mostrar el número de nodos filtrados en el panel de estado.
    [#509](https://github.com/weaveworks/scope/issues/509)

Correcciones:

*   Evite que el panel de detalles oculte los nodos al hacer clic para enfocar.
    [#495](https://github.com/weaveworks/scope/issues/495)
*   Evite que la vista radial rebote en algunas circunstancias.
    [#496](https://github.com/weaveworks/scope/issues/496)
*   Haga que el componente de seguimiento NAT sea más resistente a los fallos.
    [#506](https://github.com/weaveworks/scope/issues/506)
*   Evite que los informes duplicados lleguen a la misma aplicación.
    [#463](https://github.com/weaveworks/scope/issues/463)
*   Mejorar la consistencia de la direccionalidad de los bordes en algunos casos de uso.
    [#373](https://github.com/weaveworks/scope/issues/373)
*   Asegúrese de que la sonda, la aplicación y el contenedor se apaguen de forma limpia.
    [#424](https://github.com/weaveworks/scope/issues/424)
    [#478](https://github.com/weaveworks/scope/issues/478)

## Versión 0.7.0

Nuevas características:

*   Mostrar conexiones de corta duración en la vista de contenedores.
    [#356](https://github.com/weaveworks/scope/issues/356)
    [#447](https://github.com/weaveworks/scope/issues/447)
*   Panel de detalles:
    1.  Añadir más información:
    *   Etiquetas Docker.
        [#400](https://github.com/weaveworks/scope/pull/400)
    *   Tejer IPs/nombres de host/MAC e IP de Docker.
        [#394](https://github.com/weaveworks/scope/pull/394)
        [#396](https://github.com/weaveworks/scope/pull/396)
    *   Contexto de host/contenedor cuando es ambiguo.
        [#387](https://github.com/weaveworks/scope/pull/387)
    2.  Preséntalo de una manera más intuitiva:
    *   Muestre nombres de host en lugar de IP cuando sea posible.
        [#404](https://github.com/weaveworks/scope/pull/404)
        [#451](https://github.com/weaveworks/scope/pull/451)
    *   Combine toda la información relacionada con la conexión en una sola tabla.
        [#322](https://github.com/weaveworks/scope/issues/322)
    *   Incluya información relevante en los títulos de las tablas.
        [#387](https://github.com/weaveworks/scope/pull/387)
    *   Deje de incluir campos vacíos.
        [#370](https://github.com/weaveworks/scope/issues/370)
*   Permitir filtrar los contenedores del sistema (por ejemplo, contenedores Weave y Scope) y
    contenedores no conectados. Los contenedores del sistema se filtran de forma predeterminada.
    [#420](https://github.com/weaveworks/scope/pull/420)
    [#337](https://github.com/weaveworks/scope/issues/337)
    [#454](https://github.com/weaveworks/scope/issues/454)
    [#457](https://github.com/weaveworks/scope/issues/457)
*   Mejore la representación haciendo que las direcciones de los bordes fluyan del cliente al servidor.
    [#355](https://github.com/weaveworks/scope/pull/355)
*   Resaltar nodo seleccionado
    [#473](https://github.com/weaveworks/scope/pull/473)
*   Animar bordes durante las transciones de la interfaz de usuario
    [#445](https://github.com/weaveworks/scope/pull/445)
*   Nueva barra de estado en la parte inferior izquierda de la interfaz de usuario
    [#487](https://github.com/weaveworks/scope/pull/487)
*   Muestre más información para los pseudonodos cuando sea posible, como los procesos para no contenidos y conectados a/desde Internet.
    [#249](https://github.com/weaveworks/scope/issues/249)
    [#401](https://github.com/weaveworks/scope/pull/401)
    [#426](https://github.com/weaveworks/scope/pull/426)
*   Truncar nombres de nodo y texto en el panel de detalles.
    [#429](https://github.com/weaveworks/scope/pull/429)
    [#430](https://github.com/weaveworks/scope/pull/430)
*   Amazon ECS: Deje de mostrar nombres de contenedores destrozados, muestre en su lugar el nombre original de la definición de contenedor.
    [#456](https://github.com/weaveworks/scope/pull/456)
*   Anote procesos en contenedores con el nombre del contenedor, en el cuadro *Aplicaciones* vista.
    [#331](https://github.com/weaveworks/scope/issues/331)
*   Mejore las transiciones de gráficos entre actualizaciones.
    [#379](https://github.com/weaveworks/scope/pull/379)
*   Reducir el uso de CPU de las sondas Scope
    [#470](https://github.com/weaveworks/scope/pull/470)
    [#484](https://github.com/weaveworks/scope/pull/484)
*   Hacer que la propagación de informes sea más fiable
    [#459](https://github.com/weaveworks/scope/pull/459)
*   Interfaz de estado de Support Weave 1.1
    [#389](https://github.com/weaveworks/scope/pull/389)

Correcciones:

*   *Intentando reconectarse..* en la interfaz de usuario aunque esté conectada.
    [#392](https://github.com/weaveworks/scope/pull/392)
*   El *Aplicaciones* la vista se queda en blanco después de unos segundos.
    [#442](https://github.com/weaveworks/scope/pull/442)
*   Obtener líneas desconectadas con frecuencia en la interfaz de usuario
    [#460](https://github.com/weaveworks/scope/issues/460)
*   Pánico debido al cuerpo de solicitud de cierre
    [#480](https://github.com/weaveworks/scope/pull/480)

## Versión 0.6.0

Nuevas características:

*   Los sondeos ahora envían datos a la aplicación, en lugar de que la aplicación los extraiga.
    [#342](https://github.com/weaveworks/scope/pull/342)
*   Permitir que la sonda y la aplicación se inicien de forma independiente, a través de --no-app y
    \--banderas sin sonda.
    [#345](https://github.com/weaveworks/scope/pull/345)
*   Cerrar panel de detalles al cambiar la vista de topología.
    [#297](https://github.com/weaveworks/scope/issues/297)
*   Agregue soporte para los indicadores de estilo --probe.foo=bar, además de
    \--probe.foo bar, que ya es compatible.
    [#347](https://github.com/weaveworks/scope/pull/347)
*   Se ha agregado el encabezado X-Scope-Probe-ID para identificar sondas al enviar
    información a la aplicación.
    [#351](https://github.com/weaveworks/scope/pull/351)

Correcciones:

*   Actualizar el script de ámbito para que funcione con la versión maestra de weave, donde DNS
    se ha incrustado en el router.
    [#321](https://github.com/weaveworks/scope/issues/321)
*   Se corrigió la regresión en la que los nombres de los procesos no aparecían para Darwin
    Sondas.
    [#320](https://github.com/weaveworks/scope/pull/320)
*   Se ha corregido un error de representación que daba lugar a nodos huérfanos.
    [#339](https://github.com/weaveworks/scope/pull/339)
*   La aplicación ahora solo inicia sesión en stderr, para que coincida con la sonda.
    [#343](https://github.com/weaveworks/scope/pull/343)
*   Use rutas de acceso relativas para todas las direcciones URL de la interfaz de usuario.
    [#344](https://github.com/weaveworks/scope/pull/344)
*   Se han quitado los contenedores temporales creados por el script de ámbito.
    [#348](https://github.com/weaveworks/scope/issues/348)

Características experimentales:

*   Se agregó soporte para la detección de paquetes basada en pcap, para proporcionar ancho de banda
    información de uso. Se puede habilitar a través de la bandera --capture. Cuando
    Habilitada, la sonda supervisará los paquetes durante una parte del tiempo, y
    estimar el uso del ancho de banda. El rendimiento de la red se verá afectado si
    la captura está habilitada.
    [#317](https://github.com/weaveworks/scope/pull/317)

## Versión 0.5.0

Nuevas características:

*   Agregue toda la información de conexión en una sola tabla en los detalles
    diálogo.
    [#298](https://github.com/weaveworks/scope/pull/298)
*   Se ha cambiado el nombre de los archivos binarios a scope-app y scope-probe
    [#293](https://github.com/weaveworks/scope/pull/293)
*   Topología de contenedores de grupo solo por nombre y no por versión
    [#291](https://github.com/weaveworks/scope/issues/291)
*   Haga que la comunicación intra-alcance atraviese la red de tejido si está presente.
    [#71](https://github.com/weaveworks/scope/issues/71)

Correcciones:

*   Uso de memoria reducido
    [#266](https://github.com/weaveworks/scope/issues/266)

## Versión 0.4.0

Nuevas características:

*   Incluya la versión del kernel y el tiempo de actividad en los detalles del host.
    [#274](https://github.com/weaveworks/scope/pull/274)
*   Incluya la línea de comandos y el número de subprocesos en los detalles del proceso.
    [#272](https://github.com/weaveworks/scope/pull/272)
*   Incluye mapeo de puertos Docker, punto de entrada, uso de memoria y creación
    fecha en los detalles del contenedor.
    [#262](https://github.com/weaveworks/scope/pull/262)
*   Ordene las tablas en el panel de detalles de menos granular a más granular.
    [#261](https://github.com/weaveworks/scope/issues/261)
*   Mostrar todas las imágenes de contenedor (incluso las que no tienen conexiones activas)
    en la vista contenedores por imagen.
    [#230](https://github.com/weaveworks/scope/issues/230)
*   Producir vistas de procesos y contenedores combinando la topología de punto final con su
    topologías respectivas, de modo que los orígenes en el panel de detalles siempre
    agregar correctamente. [#228](https://github.com/weaveworks/scope/issues/228)
*   En la vista de contenedores, muestre los nodos "No contenidos" para cada host si tienen
    conexiones activas. [#127](https://github.com/weaveworks/scope/issues/127)
*   Mostrar el estado de la conexión en la interfaz de usuario.
    [#162](https://github.com/weaveworks/scope/issues/162)

Correcciones:

*   Reduzca el uso de la CPU mediante el almacenamiento en caché de /proc.
    [#284](https://github.com/weaveworks/scope/issues/284)
*   Recortar el espacio en blanco de los nombres de proceso de modo que funcione el panel de detalles
    correctamente en la vista proceso por nombre.
    [#281](https://github.com/weaveworks/scope/issues/281)
*   Direcciones de ámbito correctas en el puente Docker de modo que los procesos en diferentes
    los hosts no se muestran incorrectamente como comunicantes.
    [#264](https://github.com/weaveworks/scope/pull/264)
*   Mostrar correctamente las conexiones entre los nodos que atraviesan un puerto Docker
    cartografía. [#245](https://github.com/weaveworks/scope/issues/245)
*   Haga que el script de ámbito falle si se produce un error en la ejecución de Docker.
    [#214](https://github.com/weaveworks/scope/issues/214)
*   Impedir los nodos sobrantes en la interfaz de usuario cuando se interrumpe la conexión o se produce el ámbito
    Reinicia. [#162](https://github.com/weaveworks/scope/issues/162)

## Versión 0.3.0

*   Muestre contenedores, incluso cuando no se estén comunicando.
*   Expanda los selectores de topología más descriptivos y quite el botón de agrupación.
*   Corregir errores de representación de desbordamiento en el panel de detalles.
*   Renderiza pseudonodos con menos saturación.

## Versión 0.2.0

*   Versión inicial.
