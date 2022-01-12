<?php

namespace AppBundle\Form;

use AppBundle\Entity\Log;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class LogType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('date')
            ->add('server')
            ->add('method')
            ->add('request')
            ->add('param')
            ->add('port')
            ->add('user')
            ->add('client')
            ->add('agent')
            ->add('referer')
            ->add('status')
            ->add('substatus')
            ->add('win32')
            ->add('duration')
        ;
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class' => Log::class,
        ]);
    }
}
