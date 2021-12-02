# Microfony Framework

Microfony is a micro framework based on [Symfony](https://symfony.com/) 4.4 LTS.

To be minimal the frontend page is build with [Pure](https://purecss.io) and uses [Twig](https://twig.symfony.com) as Template Engine.

## Requirements

To check the requirements run

    bin/symfony_requirements
    
## Composer

    sudo curl -LsS https://getcomposer.org/installer -o /usr/local/bin/composer
    sudo chmod a+x /usr/local/bin/composer
    sudo composer self-update

## Installation

    git clone https://gitlab.com/typomedia/microfony.git
    cd microfony
    composer install
    php -S localhost:8000 -t web

## Developer

For [PhpStorm](https://www.jetbrains.com/phpstorm/) users install the following Plugins:

* [Symfony Plugin](https://plugins.jetbrains.com/plugin/7219-symfony-plugin)
* [PHP Annotations](https://plugins.jetbrains.com/plugin/7320-php-annotations)

php bin/console doctrine:mapping:import "App\Entity" annotation --path=src/Entity
