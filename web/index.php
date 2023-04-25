<?php

use Symfony\Component\ErrorHandler\Debug;
use Symfony\Component\HttpFoundation\Request;

ini_set('post_max_size', '512M');
ini_set('upload_max_filesize', '512M');
ini_set('max_input_time', 600);
ini_set('max_execution_time', 600);

/**
 * @var Composer\Autoload\ClassLoader $loader
 */
require __DIR__.'/../vendor/autoload.php';
Debug::enable();

$kernel = new AppKernel('dev', true);

$request = Request::createFromGlobals();
$response = $kernel->handle($request);
$response->send();
$kernel->terminate($request, $response);
