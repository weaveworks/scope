# Uso de código de proveedor en Weave Scope

Weave Scope está licenciado bajo el [Licencia de Apache 2.0](LICENSE).

Sin embargo, algunos códigos de proveedores están bajo diferentes licencias, todas ellas envían el
todo el texto de la licencia bajo el que se encuentran.

*   https://github.com/weaveworks/go-checkpoint\
    https://github.com/weaveworks/go-cleanhttp\
    https://github.com/certifi/gocertifi\
    se puede encontrar en el directorio ./vendor/, está bajo MPL-2.0.

*   Atraídos por dependencias son\
    https://github.com/hashicorp/go-version (MPL-2.0)\
    https://github.com/hashicorp/golang-lru (MPL-2.0)

*   Un archivo extraído por una dependencia está bajo CDDL:\
    ./vendor/github.com/howeyc/gopass/terminal_solaris.go

*   Los documentos de una dependencia que es atraída por una dependencia
    están bajo CC-BY 4.0:
    ./vendor/github.com/docker/go-units/

[Un archivo utilizado en las pruebas](COPYING.LGPL-3) está bajo LGPL-3, es por eso que enviamos
el texto de la licencia en este repositorio.
