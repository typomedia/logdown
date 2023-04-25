# Logdown

Logdown is a Log analyzer for IIS Logs.

## Upload Settings

In the `nginx.conf` file you have to set the following settings:

    client_max_body_size 200M

In the `fpm/php.ini` file you have to set the following settings:

    upload_max_filesize = 200M
    max_file_uploads = 40
    post_max_size = 200M
    memory_limit = 256M
    max_execution_time = 600
    max_input_time = 600

    
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
