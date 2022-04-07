<?php

namespace AppBundle\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Kernel;
use Symfony\Component\Routing\Annotation\Route;
use Typomedia\Sysinfo\Exception\UnsupportedSystemException;
use Typomedia\Sysinfo\SysinfoFactory;

/**
 * @Route("/about")
 */
class AboutController extends AbstractController
{
    /**
     * @Route("/")
     * @throws UnsupportedSystemException
     */
    public function info(): Response
    {
        $sysinfo = SysinfoFactory::create();

        $system = [
            'os' => $sysinfo->getOsRelease(),
            'cpu' => $sysinfo->getCpuModel(),
            'ram' => $sysinfo->getTotalMem(),
            'php' => $sysinfo->getPhpVersion(),
            'symfony' => Kernel::VERSION
        ];

        return $this->render('@App/sites/about.html.twig', ['system' => $system]);
    }
}
