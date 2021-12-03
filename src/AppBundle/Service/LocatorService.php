<?php

namespace AppBundle\Service;

use Symfony\Component\HttpKernel\Config\FileLocator;

class LocatorService
{
    private $fileLocator;

    public function __construct(FileLocator $fileLocator)
    {
        $this->fileLocator = $fileLocator;
    }

    public function doSth()
    {
        $resourcePath = $this->fileLocator->locate('@AppBundle');
    }
}
