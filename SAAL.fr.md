# Manifeste du SaaL

## Avant propos

Nous allons discuter ici des implications de la création d'un logiciel respectant les principes du "Software as a Library" (que nous appelerons désormais SaaL).

Le SaaL est un sous-ensemble de ce que l'on appel plus communément "framework" dans le développement d'application.

La définition du terme de SaaL a pour but de mettre en évidence un type particulier d'application afin de pouvoir généraliser cette pratique en dehors des cadres habituels dans lesquelles elle est conçue. 

## 1. But

Le SaaL a pour but d'offrir un logiciel que l'on peut configurer et étendre par le code AVANT la compilation.

Il est une alternative pour les développeurs aux logiciels hautement configurables.

En effet, si une interface web pour une application hautement configurable est beaucoup plus intuitive et facile d'utilisation que du code pour un profil non technique, elle constitue pour le développeur une perte de temps très importante. Il n'y a, pour un développeur, pas de meilleur interface que son IDE, pas de meilleur vérification de la cohérence de l'application que le compilateur d'un langage au typage statique, pas de meilleure façon de debugger que son debugger, etc.

C'est un problème qui s'observe notamment dans l'écosystème d'outils de développement et de Devops. Souvent, on est obligé de faire trois à cinq cliques sur une interface web pour aller définir une variable, puis l'on doit utiliser cette variable dans une partie d'un formulaire où l'on définit un script sans pouvoir s'assurer de ne pas faire de typo, on duplique beaucoup de configuration, l'outil ne nous donne pas forcément accès à l'environnement où il est exécuté, oblige de passer par une abstraction opaque pour accéder à cet environnement, etc.

Alors que le code notre IDE nous offre de l'autocomplétion, le code nous permet de factoriser les parties dupliquées, et l'environnement d'exécution du code est entièrement sous le contrôle du développeur. Le code nous permet également de tester l'application avant de la déployer, nous assurant de ne pas découvrir des soucis au moment de la mise en production.

Au délà de cela, les outils utilisés, du fait qu'ils cherchent à être hautement configurables sans pour autant donner la liberté offerte par le code, sont souvent très lourds et proposent un nombre de fonctionnalités bien trop important pour des cas d'utilisation simples.

Il est également à prendre en compte le fait que ces outils nécessitent un apprentissage de compétences non généralisables : les connaissances avancées sur un outil ne sont pas transposables à un autre.  Toutefois, le travail de devops ou développeur fullstack demande un panel de compétence si important que la connaissance spécifique à un outil n'étant pas un standard est un investissement difficilement rentable. 

Le SaaL est donc la proposition permettant aux développeurs de déployer des applications n'étant que des variantes de choses déjà existantes sans avoir à souffrir de l'expérience des applications hautement configurables et sans avoir pour autant à réinventer la roue.

De plus, étant donné la généralisation de l'utilisation de l'IA générative pour l'aide au développement, les solutions habituellement proposées en no-code pour satisfaire les besoins d'utilisateurs non techniques peuvent se permettre de passer en full-code si les questions d'implémentation de la sécurité et de gestion de l'état de l'application sont gérées par le SaaL.


## 2. Exemples

### 2.1 Logiciels SaaL ou adjacents au SaaL

 * Docusaurus est très proche du SaaL, dans le sens où il s'agit d'un outil full-code distribué sous forme de librairie totalement extensible mais dont la principale fonctionnalité peut s'accomplir simplement en appelant les fonctions disponibles dans le code

 * Hugo, très proche de Docusaurus, peut également être vu comme un SaaL

 * Laravel propose des mécanismes proches du SaaL, notamment avec son système d'authentification intégré par défaut. On pourrait le voir comme un SaaL de conception d'API Rest avec authentification.

 * NextJS propose des mécanismes proches du SaaL, dans le sens où il orchestre à lui seul la totalité du cycle de vie d'une application web.

### 2.2 Logiciels non SaaL

 * Jenkins: il propose principalement une interface via un GUI, et une configuration de son fonctionnement via cette interface
 * React: si react peut répondre à certains critères du SaaL, il n'est pas seul suffisant pour être utilisé dans le contexte d'une application web : on a également besoin d'un serveur avec une base de données.

### 2.3 Domaines qui gagneraient à avoir un SaaL

 * Les domaines du CI/CD demandent généralement d'exécuter une suite de commandes bash et des appels à des services externes, en gérant les cas d'erreur et en produisant un reporting de qualité sur le déroulement du processus. Souvent, il y a besoin d'utiliser des services comme Docker, SSH, Git, dont la configuration dépend de l'environnement d'exécution du processus. Il y a donc un réel intérêt à fournir un langage de programmation complet et un accès direct à l'environnement d'exécution pour ces tâches.

 * Les ERPs et LMS peuvent demander des extensions très spécifiques à certains domaines, peuvent demander énormément de customisation via des formulaires, et peuvent facilement devenir d'insupportables usines à gaz s'ils sont mal configurés. Le fait de devoir s'interfacer avec énormément de type de données (fichiers excel, bases de données, API diverses) va de toutes façon souvent demander l'intervention d'un développeur pour le maintenir. Mieux vaut donc avoir un serveur auto-configuré avec une structure de données par défaut sur laquelle on peut venir ajouter des fonctionnalités  

## 3. Définition

### 3.1. le SaaL est avant tout une librairie

**Axiome** 

Le SaaL *DOIT* être proposé en tant que librairie pour un langage de programmation. Cette librairie *PEUT* nécessiter des fichiers de configuration. Si c'est la cas, ces fichiers *DOIVENT* être disposés de façon à ce qu'ils puissent être accessible dans le même dépôt git que l'application, même s'ils sont ensuite ajoutés au .gitignore. De la même façon, le SaaL *DOIT* fournir un exécutable capable de fournir les fichiers de configuration nécessaires avec des valeurs par défaut. 

**Lemme**

Le SaaL *PEUT* fournir un CLI permettant de générer du code et des fichiers de configuration, voir plus. Cependant le SaaL *DEVRAIT* éviter de rendre ce CLI trop complexe.

**Développement**



### 3.2. le SaaL est l'orchéstrateur unique de l'application

**Axiome**

Pour toute opération lié au domaine du logiciel, un SaaL *DOIT* offrir une série de fonctions dans lesquelles l'utilisateur devrait pouvoir écrire ou appeler la totalité de son code, fonction main mise à part. Toute variable déclarée dans la fonction main doit être le résultat d'un appel d'une fonction du SaaL.


**Développement**

Imaginons un SaaL ayant pour but de développer une API REST, et comparons là avec un framework web comme Gin. Gin propose de définir les routes de cette façon:

```go
func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.Conf.CORSES},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	//Use the custom renderer to serve the templ compiled files
	router.HTMLRender = &config.TemplRender{}

	//Add static routes
	router.Static("/css", "../resources/static/css")
	router.Static("/js", "../resources/static/js")
	router.Static("/images", "../resources/static/images")
	router.StaticFile("/favicon.ico", "../resources/static/favicon.ico")

	//Create the groupes
	baseGroup := router.Group("/")
	baseGroup.Use(middleware.LocaleMiddleware())
	baseGroup.Use(middleware.UserInfoMiddleware())

	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.LocaleMiddleware())
	apiGroup.Use(middleware.Authenticate)
	apiGroup.Use(middleware.PreventUnverified)

	authButNotVerified := router.Group("/")
	authButNotVerified.Use(middleware.LocaleMiddleware())
	authButNotVerified.Use(middleware.Authenticate)

    baseGroup.POST("/signup", signUp) // signUp est une fonction définie dans un autre fichier
    baseGroup.POST("/login", login) // login est une fonction définie dans un autre fichier
    baseGroup.GET("/activation/:uuid", verify) // verify est une fonction définie dans un autre fichier
    baseGroup.POST("/password/ask-new", forgottenPassword) // forgottenPassword est une fonction définie dans un autre fichier

	router.Run("0.0.0.0:42069")
}

```

Pour la création des routes, gin respecte les principes du SaaL : toutes les variables définies dans la fonction main sont des retours de fonctions internes à Gin. Toutefois, Gin n'est pas un SaaL pour les API Rest : en effet, pour créer ma base de données, je vais devoir utiliser des outils externes à Gin, que je NE PEUX PAS configurer en utilisant des outils internes à Gin. C'est une des différences entre le SaaL est le framework : le SaaL NE PEUT PAS être SaaL s'il ne peut pas remplir son objectif sans devoir être orchestré avec d'autres fonctions, le framework PEUT être framework en étant orchestré avec d'autres fonctions.

Pour que Gin puisse être un SaaL, il aurait par exemple fallu pouvoir écrire cela:

```go
func main() {
    server := gin.NewServer()
    server.NewSQLiteDatabase("./resources/database.sql")
    router := server.Router("/api/v1/")
    router.Authed()
    router.GET("/test", func(c *gin.Context) {
        test := c.db.Execute("SELECT * FROM test")
        c.JSON(http.StatusOK, test)
    })
	server.Run("0.0.0.0:42069")
}
```

À noter que la création d'une API Rest générique n'est pas forcément un bon exemple d'application ayant vocation à être un SaaL. Car les applications que l'on peut créer à partir de cette idée varient énormément : Twitter, un ERP, un chatbot IA, etc.

Toutefois, on peut imaginer un SaaL pour un type de site en particulier. Pour la documentation, par exemple, Docusaurus répond presque en tous points à la définition du SaaL. On pourrait également en imaginer un pour les blogs, les ERPs, les LMS, etc.

### 3.3 Le SaaL gère l'état de l'application de façon opaque.

**Axiome**

Un SaaL *DEVRAIT* permettre aux utilisateurs de passer des comportements aux interfaces du SaaL en respectant le principe de **Locality of Behaviour** (LoB). Une fonction passée à une interface du SaaL *NE DEVRAIT PAS* changer les règles prédéfinies sur la façon dont les autres fonctions sont exécutées.

### 3.4 Le SaaL fournit des utilitaires pour les opérations les plus communes de son domaine

**Axiome**

Si le SaaL





